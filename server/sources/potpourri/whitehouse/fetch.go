package whitehouse

import (
	"encoding/xml"
	"log"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/types"
	"git.nunosempere.com/NunoSempere/news/lib/web"
)

const feedURL = "https://www.whitehouse.gov/presidential-actions/feed/"

// "https://www.whitehouse.gov/briefing-room/statements-releases/feed/index.xml"
//"https://www.whitehouse.gov/blog/feed/"
// https://www.federalregister.gov/api/v1/documents.rss?conditions%5Bpresident%5D%5B%5D=donald-trump&conditions%5Bpresidential_document_type%5D%5B%5D=executive_order&conditions%5Btype%5D%5B%5D=PRESDOCU
// https://www.reddit.com/r/InoReader/comments/1i6nopb/white_house_rss_feed/

func FetchFeed() ([]types.Source, error) {
	log.Printf("Fetching White House feed: %s", feedURL)

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
			Date:   pubDate,
			Origin: "White House",
		})
	}

	return sources, nil
}
