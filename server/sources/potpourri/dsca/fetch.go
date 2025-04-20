package dsca

import (
	"encoding/xml"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"git.nunosempere.com/NunoSempere/news/lib/web"
	"log"
	"time"
)

const feedURL = "https://www.dsca.mil/DesktopModules/ArticleCS/RSS.ashx?ContentType=700&Site=1509&isdashboardselected=0&max=20"

func FetchFeed() ([]types.Source, error) {
	log.Printf("Fetching DSCA feed: %s", feedURL)
	
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
		// Try to parse the date
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
		if time.Since(pubDate) > 24*time.Hour {
			continue
		}

		sources = append(sources, types.Source{
			Title:  item.Title,
			Link:   item.Link,
			Date:   pubDate.Format(time.RFC3339),
			Origin: "DSCA",
		})
	}

	return sources, nil
}