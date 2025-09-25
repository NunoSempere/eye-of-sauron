package main

import (
	"strings"

	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

// FilterAndExpandSource processes a xAI tweet article through various filters,
// expands its content (via summarization and importance check),
// and returns an ExpandedSource and a boolean indicating if it passes thresholds.
func FilterAndExpandSource(source types.Source, openai_key string, database_url string) (types.ExpandedSource, bool) {
	// Initialize expanded source with basic info
	// Note: The article content is currently stored in the Title field from fetch.go
	// We'll extract a proper title and use the content as the summary base
	articleContent := source.Title
	properTitle := extractTitleFromContent(articleContent)

	es := types.ExpandedSource{
		Title:  properTitle,
		Link:   source.Link,
		Date:   source.Date,
		Origin: source.Origin,
	}

	// Apply filters - skip some that don't make sense for tweet collections
	fs := []types.Filter{
		filters.IsFreshFilter(),
		filters.IsDupeFilter(database_url),
		filters.CleanTitleFilter(),
		// Use the article content directly as summary instead of extracting from web
		createDirectSummaryFilter(articleContent),
		filters.CheckImportanceFilter(openai_key),
	}
	es, ok := filters.ApplyFilters(es, fs)
	if !ok {
		return es, false
	}
	return es, true
}

// extractTitleFromContent extracts a clean title from the article content
func extractTitleFromContent(content string) string {
	// Look for the first line which should be the title
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && strings.HasPrefix(trimmed, "#") {
			// Remove the markdown header prefix
			title := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			return title
		}
	}
	return "Tweet Collection" // fallback
}

// createDirectSummaryFilter creates a filter that uses the article content directly as summary
func createDirectSummaryFilter(content string) types.Filter {
	return func(es types.ExpandedSource) (types.ExpandedSource, bool) {
		// Use the full article content as the summary
		es.Summary = content
		return es, true
	}
}

