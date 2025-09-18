package main

import (
	"log"

	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/readability"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

// FilterAndExpandSource processes a {{SOURCE_NAME}} source through various filters,
// expands its content (via summarization and importance check),
// and returns an ExpandedSource and a boolean indicating if it passes thresholds.
func FilterAndExpandSource(source types.Source, openai_key string, database_url string) (types.ExpandedSource, bool) {
	// Use standard processing pipeline with optional title extraction
	es := types.ExpandedSource{
		Title: source.Title,
		Link:  source.Link,
		Date:  source.Date,
		Origin: source.Origin,
	}

	// Apply standard filters
	filters_list := filters.StandardFilterPipeline(database_url)
	es, ok := filters.ApplyFilters(es, filters_list)
	if !ok {
		return es, false
	}

	// Try to get a better title from the source HTML
	if title := readability.ExtractTitle(source.Link); title != "" {
		es.Title = title
		log.Printf("Found title from HTML: %s", title)
		// Clean the extracted title
		es.Title = filters.CleanTitle(es.Title)
	}

	// Extract content and summarize
	es, ok = filters.ExtractContentAndSummarize(es, openai_key)
	if !ok {
		return es, false
	}

	// Check importance
	return filters.CheckImportance(es, openai_key)
}
