package main

import (
	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

// FilterAndExpandSource processes a {{SOURCE_NAME}} source through various filters,
// expands its content (via summarization and importance check),
// and returns an ExpandedSource and a boolean indicating if it passes thresholds.
func FilterAndExpandSource(source types.Source, openai_key string, database_url string) (types.ExpandedSource, bool) {
	// Use standard processing pipeline with optional title extraction
	es := types.ExpandedSource{
		Title:  source.Title,
		Link:   source.Link,
		Date:   source.Date,
		Origin: source.Origin,
	}

	fs := []types.Filter{
		filters.IsFreshFilter(),
		filters.IsDupeFilter(database_url),
		filters.IsGoodHostFilter(),
		filters.CleanTitleFilter(),
		filters.ExtractSummaryFilter(openai_key),
		filters.CheckImportanceFilter(openai_key),
	}
	es, ok := filters.ApplyFilters(es, fs)

	return es, ok

}
