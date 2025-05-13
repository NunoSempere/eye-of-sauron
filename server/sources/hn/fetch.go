package main

import (
	"encoding/json"
	"fmt"
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

func FetchFeed() ([]HNHit, error) {
	// Calculate time range for the last hour
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour) // give time to accumulate upvotes
	
	var allSources []HNHit
	currentPage := 0
	
	// Fetch first page to get pagination info
	response, err := fetchPage(currentPage, twoHoursAgo.Unix(), oneHourAgo.Unix())
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
		allSources = append(allSources, response.Hits...)
	}

	log.Printf("Processed %d valid HN stories", len(allSources))
	return allSources, nil
}

