package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchAggregatesResults(t *testing.T) {
	// Stub Google server
	gSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/url?q=http://g.com&amp;sa=U"><h3>GTitle</h3></a>`))
	}))
	defer gSrv.Close()

	bSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<li class="b_algo"><h2><a href="http://b.com">BTitle</a></h2></li>`))
	}))
	defer bSrv.Close()

	baSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<h3 class="t"><a href="http://baidu.com">BaTitle</a></h3>`))
	}))
	defer baSrv.Close()

	// Override search URLs
	oldGoogleURL := googleURL
	oldBingURL := bingURL
	oldBaiduURL := baiduURL
	googleURL = gSrv.URL + "?q=%s"
	bingURL = bSrv.URL + "?q=%s"
	baiduURL = baSrv.URL + "?wd=%s"
	defer func() {
		googleURL = oldGoogleURL
		bingURL = oldBingURL
		baiduURL = oldBaiduURL
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/search", searchHandler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/search?q=test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var agg AggregatedResults
	if err := json.NewDecoder(resp.Body).Decode(&agg); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(agg.Google) != 1 || agg.Google[0].Title != "GTitle" || agg.Google[0].URL != "http://g.com" {
		t.Errorf("google results mismatch: %+v", agg.Google)
	}
	if len(agg.Bing) != 1 || agg.Bing[0].Title != "BTitle" || agg.Bing[0].URL != "http://b.com" {
		t.Errorf("bing results mismatch: %+v", agg.Bing)
	}
	if len(agg.Baidu) != 1 || agg.Baidu[0].Title != "BaTitle" || agg.Baidu[0].URL != "http://baidu.com" {
		t.Errorf("baidu results mismatch: %+v", agg.Baidu)
	}
}
