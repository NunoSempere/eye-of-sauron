package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/types"
)

// TODO: Define any custom types needed for your source's API/RSS response
type APIResponse struct {
	// Add fields based on your source's response format
}

type RSSResponse struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Items []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// FetchSources retrieves sources from {{SOURCE_NAME}}
// TODO: Implement your source-specific fetching logic
func FetchSources() ([]types.Source, error) {
	var sources []types.Source

	// Example for RSS feed:
	// sources, err := fetchFromRSS("{{RSS_URL}}")

	// Example for API:
	// sources, err := fetchFromAPI("{{API_URL}}")

	// Example for web scraping:
	// sources, err := fetchFromWebpage("{{WEBPAGE_URL}}")

	// TODO: Replace with actual implementation
	return sources, nil
}

// fetchFromRSS fetches sources from an RSS feed
func fetchFromRSS(url string) ([]types.Source, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rss RSSResponse
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, err
	}

	var sources []types.Source
	for _, item := range rss.Channel.Items {
		// TODO: Parse date in a different format if needed
		date, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			date = time.Now()
		}

		source := types.Source{
			Title: item.Title,
			Link:  item.Link,
			Date:  date,
		}
		sources = append(sources, source)
	}

	return sources, nil
}

// fetchFromAPI fetches sources from a JSON API
func fetchFromAPI(url string) ([]types.Source, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	var sources []types.Source
	// TODO: Convert API response to types.Source slice

	return sources, nil
}

func doSmth(_ interface{}) {

}

// fetchFromWebpage scrapes sources from a webpage
func fetchFromWebpage(url string) ([]types.Source, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	doSmth(content)

	// TODO: Parse HTML content to extract sources
	// This might involve regex, HTML parsing, or other extraction methods
	var sources []types.Source

	return sources, nil
}
