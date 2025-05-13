package main

import (
	"encoding/json"
	"fmt"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"git.nunosempere.com/NunoSempere/news/lib/web"
	"log"
	"time"
)

func fetchPage(page int, startTime, endTime int64) (*HNResponse, error) {
	url := fmt.Sprintf("http://hn.algolia.com/api/v1/search_by_date?tags=story&numericFilters=created_at_i>%d,created_at_i<%d&page=%d",
		startTime,
		endTime,
		page)
	
	log.Printf("Fetching HN feed page %d: %s", page, url)
	
	bytes, err := web.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching HN feed: %v", err)
	}

	var response HNResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling HN response: %v", err)
	}

	return &response, nil
}

func FetchFeed() ([]types.Source, error) {
	// Calculate time range for the last hour
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	
	var allSources []types.Source
	currentPage := 0
	
	// Fetch first page to get pagination info
	response, err := fetchPage(currentPage, oneHourAgo.Unix(), now.Unix())
	if err != nil {
		return nil, err
	}

	log.Printf("Found %d HN stories across %d pages", response.NbHits, response.NbPages)

	// Process all pages
	for currentPage = 0; currentPage < response.NbPages; currentPage++ {
		if currentPage > 0 {
			response, err = fetchPage(currentPage, oneHourAgo.Unix(), now.Unix())
			if err != nil {
				log.Printf("Error fetching page %d: %v", currentPage, err)
				continue
			}
		}

		for _, hit := range response.Hits {
			// Skip items without URLs and with low engagement
			if hit.URL == "" && hit.StoryText == "" {
				continue
			}

			// Skip low engagement posts (less than 2 points or comments). May come back to this.
			if hit.Points < 2 && hit.NumComments < 2 {
				continue
			}

			// Parse the created_at time
			createdAt, err := time.Parse(time.RFC3339, hit.CreatedAt)
			if err != nil {
				log.Printf("Could not parse date '%s', using current time", hit.CreatedAt)
				createdAt = time.Now()
			}

			source := types.Source{
				Title:  hit.Title,
				Link:   hit.URL,
				Date:   createdAt.Format(time.RFC3339),
				Origin: "HackerNews",
			}

			allSources = append(allSources, source)
		}
	}

	log.Printf("Processed %d valid HN stories", len(allSources))
	return allSources, nil
}
