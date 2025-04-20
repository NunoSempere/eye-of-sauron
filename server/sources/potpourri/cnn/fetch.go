package cnn

import (
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"git.nunosempere.com/NunoSempere/news/lib/web"
	"log"
	"time"
  "encoding/xml"
)

var feeds = map[string]string{
	"top":        "http://rss.cnn.com/rss/cnn_topstories.rss",
	"world":      "http://rss.cnn.com/rss/cnn_world.rss",
	"us":         "http://rss.cnn.com/rss/cnn_us.rss",
	"politics":   "http://rss.cnn.com/rss/cnn_allpolitics.rss",
	"technology": "http://rss.cnn.com/rss/cnn_tech.rss",
	"latest":     "http://rss.cnn.com/rss/cnn_latest.rss",
}

func FetchFeed(feedURL string) ([]types.Source, error) {
	log.Printf("Fetching CNN feed: %s", feedURL)
	
	xml_bytes, err := web.Get(feedURL)
	if err != nil {
		return nil, err
	}

	var feed RSSFeed
	err = xml.Unmarshal(xml_bytes, &feed)
	if err != nil {
		log.Printf("Error unmarshaling XML: %v\n", err)
		return nil, err
	}

	var sources []types.Source
	for _, item := range feed.Channel.Items {
		// Parse the date to ensure it's fresh
		pubDate, err := time.Parse(time.RFC1123, item.PubDate)
		if err != nil {
			log.Printf("Error parsing date %s: %v", item.PubDate, err)
			continue
		}

		// Only include articles from the last 24 hours
		if time.Since(pubDate) > 24*time.Hour {
			continue
		}

		sources = append(sources, types.Source{
			Title: item.Title,
			Link:  item.Link,
			Date:  pubDate.Format(time.RFC3339),
		})
	}

	return sources, nil
}

func FetchAllFeeds() ([]types.Source, error) {
	var allSources []types.Source
	
	for feedName, feedURL := range feeds {
		log.Printf("Processing CNN %s feed", feedName)
		sources, err := FetchFeed(feedURL)
		if err != nil {
			log.Printf("Error fetching CNN %s feed: %v", feedName, err)
			continue
		}
		allSources = append(allSources, sources...)
	}

	return allSources, nil
}
