package cnn

import (
	"encoding/xml"
	"log"
	"strings"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/types"
	"git.nunosempere.com/NunoSempere/news/lib/web"
)

var feeds = map[string]string{
	"top":        "http://rss.cnn.com/rss/cnn_topstories.rss",
	"world":      "http://rss.cnn.com/rss/cnn_world.rss",
	"us":         "http://rss.cnn.com/rss/cnn_us.rss",
	"politics":   "http://rss.cnn.com/rss/cnn_allpolitics.rss",
	"technology": "http://rss.cnn.com/rss/cnn_tech.rss",
	"latest":     "http://rss.cnn.com/rss/cnn_latest.rss",
}

func FetchFeed(feedName string, feedURL string) ([]types.Source, error) {
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
		// Skip podcast content
		if strings.Contains(strings.ToLower(item.Link), "/audio/") ||
			strings.Contains(strings.ToLower(item.Link), "podcast") ||
			strings.Contains(strings.ToLower(item.Title), "podcast") {
			continue
		}

		// If PubDate is empty, use current time
		if item.PubDate == "" {
			sources = append(sources, types.Source{
				Title:  item.Title,
				Link:   item.Link,
				Date:   time.Now(),
				Origin: "CNN/" + feedName,
			})
			continue
		}

		// Try multiple date formats
		var pubDate time.Time
		formats := []string{
			time.RFC1123,
			time.RFC1123Z,
			"Mon, 02 Jan 2006 15:04:05 MST",
			"Mon, 2 Jan 2006 15:04:05 MST",
			"Mon, 02 Jan 2006 15:04:05 -0700",
		}

		parsed := false
		for _, format := range formats {
			if pd, err := time.Parse(format, item.PubDate); err == nil {
				pubDate = pd
				parsed = true
				break
			}
		}

		if !parsed {
			log.Printf("Could not parse date '%s' with any known format, using current time", item.PubDate)
			pubDate = time.Now()
		}

		// Skip articles older than 24 hours
		// We use > here because time.Since(pubDate) returns the duration since publication
		// If this duration is greater than 24 hours, the article is too old
		// Using < would skip articles less than 24 hours old, which is the opposite of what we want
		if time.Since(pubDate) > 24*time.Hour {
			continue
		}

		sources = append(sources, types.Source{
			Title:  item.Title,
			Link:   item.Link,
			Date:   pubDate,
			Origin: "CNN/" + feedName,
		})
	}

	return sources, nil
}

func FetchAllFeeds() ([]types.Source, error) {
	var allSources []types.Source

	for feedName, feedURL := range feeds {
		log.Printf("Processing CNN %s feed", feedName)
		sources, err := FetchFeed(feedName, feedURL)
		if err != nil {
			log.Printf("Error fetching CNN %s feed: %v", feedName, err)
			continue
		}
		allSources = append(allSources, sources...)
	}

	return allSources, nil
}
