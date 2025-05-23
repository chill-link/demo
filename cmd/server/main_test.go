package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchHandler(t *testing.T) {
	// stub search engine servers
	googleSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/url?q=http://google.com&amp;sa=U"><h3>GTitle</h3></a>`))
	}))
	defer googleSrv.Close()

	bingSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<li class="b_algo"><h2><a href="http://bing.com">BTitle</a></h2>`))
	}))
	defer bingSrv.Close()

	baiduSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<h3 class="t"><a href="http://baidu.com">BaiduTitle</a></h3>`))
	}))
	defer baiduSrv.Close()

	googleSearchURL = googleSrv.URL + "?q=%s"
	bingSearchURL = bingSrv.URL + "?q=%s"
	baiduSearchURL = baiduSrv.URL + "?wd=%s"

	server := httptest.NewServer(setupServer())
	defer server.Close()

	resp, err := http.Get(server.URL + "/search?q=test")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var result AggregatedResults
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if len(result.Google) != 1 || result.Google[0].URL != "http://google.com" {
		t.Fatalf("google results unexpected: %+v", result.Google)
	}
	if len(result.Bing) != 1 || result.Bing[0].URL != "http://bing.com" {
		t.Fatalf("bing results unexpected: %+v", result.Bing)
	}
	if len(result.Baidu) != 1 || result.Baidu[0].URL != "http://baidu.com" {
		t.Fatalf("baidu results unexpected: %+v", result.Baidu)
	}
}
