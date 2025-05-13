package hn

import (
	"encoding/json"
	"fmt"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"git.nunosempere.com/NunoSempere/news/lib/web"
	"log"
	"time"
)

func FetchFeed() ([]types.Source, error) {
	// Calculate time range for the last hour
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	
	// Create the API URL with the time filter
	url := fmt.Sprintf("http://hn.algolia.com/api/v1/search_by_date?tags=story&numericFilters=created_at_i>%d,created_at_i<%d",
		oneHourAgo.Unix(),
		now.Unix())
	
	log.Printf("Fetching HN feed: %s", url)
	
	// Fetch the data
	bytes, err := web.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching HN feed: %v", err)
	}

	// Parse the response
	var response HNResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling HN response: %v", err)
	}

	log.Printf("Found %d HN stories", len(response.Hits))

	var sources []types.Source
	for _, hit := range response.Hits {
		// Skip items without URLs
		if hit.URL == "" {
			continue
		}

		// Parse the created_at time
		createdAt, err := time.Parse(time.RFC3339, hit.CreatedAt)
		if err != nil {
			log.Printf("Could not parse date '%s', using current time", hit.CreatedAt)
			createdAt = time.Now()
		}

		sources = append(sources, types.Source{
			Title:  hit.Title,
			Link:   hit.URL,
			Date:   createdAt.Format(time.RFC3339),
			Origin: "HackerNews",
		})
	}

	return sources, nil
}
