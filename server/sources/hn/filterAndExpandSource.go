package main

import (
	"log"
	"strings"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

func FilterAndExpandSource(source HNHit, openai_key string, database_url string) (types.ExpandedSource, bool) {
	// Parse the created_at time
	createdAt, err := time.Parse(time.RFC3339, source.CreatedAt)
	if err != nil {
		log.Printf("Could not parse date '%s', using current time", source.CreatedAt)
		createdAt = time.Now()
	}

	// Initialize expanded source
	es := types.ExpandedSource{
		Title:  source.Title,
		Link:   source.URL,
		Date:   createdAt,
		Origin: "HackerNews",
	}

	// HN-specific pre-filters
	if source.URL == "" && source.StoryText == "" {
		log.Printf("No url or text")
		return es, false
	}
	if source.Points < 2 && source.NumComments < 2 {
		log.Printf("< 2 points and < 2comments")
		return es, false
	}
	if startsWithAny(source.Title, []string{"Ask HN:", "Launch HN:", "Show HN:"}) {
		log.Printf("Ask/Launch/Show HN")
		return es, false
	}

	// Apply standard filters
	fs := []types.Filter{
		filters.IsFreshFilter(),
		filters.IsDupeFilter(database_url),
		filters.IsGoodHostFilter(),
		filters.CleanTitleFilter(),
	}
	es, ok := filters.ApplyFilters(es, fs)
	if !ok {
		return es, false
	}

	// Custom content handling for HN
	if len(source.StoryText) > 100 {
		// Use story text directly if substantial
		es.Summary = source.StoryText
	} else {
		// Extract and summarize content
		es, ok = filters.ExtractContentAndSummarize(es, openai_key)
		if !ok {
			return es, false
		}
	}

	// Check importance
	es, ok = filters.CheckImportance(es, openai_key)
	if !ok {
		return es, false
	}

	// HN-specific importance boost
	if strings.Contains(source.Title, "Saudi Arabia") {
		es.ImportanceBool = true
		es.ImportanceReasoning = "Contains keyword"
	}

	return es, es.ImportanceBool
}

func startsWithAny(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
