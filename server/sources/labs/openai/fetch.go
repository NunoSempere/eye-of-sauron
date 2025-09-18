package main

import (
	"encoding/xml"
	"io"
	"net/http"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/types"
)

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

// FetchSources retrieves sources from OpenAI news RSS feed
func FetchSources() ([]types.Source, error) {
	return fetchFromRSS("https://openai.com/news/rss.xml")
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
		// Parse date from RSS format to RFC3339
		// `time.RFC1123` 
		var dateStr string
		date, err := time.Parse(time.RFC1123, item.PubDate)
		if err != nil {
				dateStr = date.Format(time.RFC3339)
		} else {
			dateStr = time.Now().Format(time.RFC3339)
		}
		
		source := types.Source{
			Title: item.Title,
			Link:  item.Link,
			Date:  dateStr,
		}
		sources = append(sources, source)
	}

	return sources, nil
}



