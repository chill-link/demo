package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sync"
)

var (
	googleSearchURL = "https://www.google.com/search?q=%s"
	bingSearchURL   = "https://www.bing.com/search?q=%s"
	baiduSearchURL  = "https://www.baidu.com/s?wd=%s"
	httpClient      = &http.Client{}
)

type SearchResult struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type AggregatedResults struct {
	Google []SearchResult `json:"google"`
	Bing   []SearchResult `json:"bing"`
	Baidu  []SearchResult `json:"baidu"`
}

func main() {
	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", setupServer()))
}

func setupServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/search", searchHandler)
	return mux
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	// Allow cross-origin requests so the API can be called from the
	// static page when it is served from a different origin.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "missing query", http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	results := AggregatedResults{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		res, err := fetchGoogle(query)
		if err == nil {
			results.Google = res
		}
	}()
	go func() {
		defer wg.Done()
		res, err := fetchBing(query)
		if err == nil {
			results.Bing = res
		}
	}()
	go func() {
		defer wg.Done()
		res, err := fetchBaidu(query)
		if err == nil {
			results.Baidu = res
		}
	}()
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func fetchGoogle(query string) ([]SearchResult, error) {
	url := fmt.Sprintf(googleSearchURL, query)
	body, err := fetchHTML(url)
	if err != nil {
		return nil, err
	}
	return parseResults(body, `href="/url\?q=([^&]+)&amp;[^"]*"[^>]*>(?:<h3[^>]*>)?(.*?)(?:</h3>)?</a>`)
}

func fetchBing(query string) ([]SearchResult, error) {
	url := fmt.Sprintf(bingSearchURL, query)
	body, err := fetchHTML(url)
	if err != nil {
		return nil, err
	}
	return parseResults(body, `<li class="b_algo"><h2><a href="([^"]+)"[^>]*>(.*?)</a>`)
}

func fetchBaidu(query string) ([]SearchResult, error) {
	url := fmt.Sprintf(baiduSearchURL, query)
	body, err := fetchHTML(url)
	if err != nil {
		return nil, err
	}
	return parseResults(body, `<h3 class="t"><a href="([^"]+)"[^>]*>(.*?)</a>`)
}

func fetchHTML(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseResults(body string, pattern string) ([]SearchResult, error) {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(body, -1)
	results := []SearchResult{}
	for _, m := range matches {
		if len(m) >= 3 {
			results = append(results, SearchResult{Title: stripTags(m[2]), URL: m[1]})
		}
	}
	if len(results) > 5 {
		results = results[:5]
	}
	return results, nil
}

func stripTags(s string) string {
	re := regexp.MustCompile("<[^>]*>")
	return re.ReplaceAllString(s, "")
}
