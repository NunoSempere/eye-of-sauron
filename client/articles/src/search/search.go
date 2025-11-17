package search

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type Search struct {
	query    string
	results  []Result
	selected int
}

type Result struct {
	Title string
	URL   string
}

func New(query string) (*Search, error) {
	s := &Search{
		query: query,
	}
	
	if err := s.fetchResults(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Search) fetchResults() error {
	bravekey := os.Getenv("BRAVE_KEY")
	searchURL := fmt.Sprintf("https://api.search.brave.com/res/v1/web/search?q=%s", 
		url.QueryEscape(s.query))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Subscription-Token", bravekey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	var braveResp BraveSearchResponse
	if err := json.Unmarshal(body, &braveResp); err != nil {
		return err
	}

	// Convert API results to our Result type
	s.results = make([]Result, len(braveResp.Web.Results))
	for i, r := range braveResp.Web.Results {
		s.results[i] = Result{
			Title: r.Title,
			URL:   r.URL,
		}
	}

	return nil
}

// Getter methods for integration
func (s *Search) GetQuery() string {
	return s.query
}

func (s *Search) GetResults() []Result {
	return s.results
}

func (s *Search) GetSelected() int {
	return s.selected
}

func (s *Search) GetSelectedResult() *Result {
	if s.selected >= 0 && s.selected < len(s.results) {
		return &s.results[s.selected]
	}
	return nil
}

// Navigation methods
func (s *Search) SelectPrevious() {
	if s.selected > 0 {
		s.selected--
	}
}

func (s *Search) SelectNext() {
	if s.selected < len(s.results)-1 {
		s.selected++
	}
}
