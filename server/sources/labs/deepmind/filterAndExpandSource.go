package main

import (
	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

// FilterAndExpandSource processes a DeepMind news source through various filters,
// expands its content (via summarization and importance check),
// and returns an ExpandedSource and a boolean indicating if it passes thresholds.
func FilterAndExpandSource(source types.Source, openai_key string, database_url string) (types.ExpandedSource, bool) {
	// Initialize expanded source with basic info
	es := types.ExpandedSource{
		Title: source.Title,
		Link:  source.Link,
		Date:  source.Date,
	}

	// TODO: check for freshness

	fs := []types.Filter{
		filters.IsFreshFilter(),
		filters.IsDupeFilter(database_url),
		filters.CleanTitleFilter(),
		filters.ExtractSummaryFilter(openai_key),
		filters.CheckImportanceFilter(openai_key),
	}
	es, ok := filters.ApplyFilters(es, fs)
	if !ok {
		return es, false
	}
	return es, true
}