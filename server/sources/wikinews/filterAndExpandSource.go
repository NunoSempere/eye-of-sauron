package main

import (
	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

// FilterAndExpandSource processes a wikinews source through various filters,
// expands its content (via summarization and importance check),
// and returns an ExpandedSource and a boolean indicating if it passes thresholds.
func FilterAndExpandSource(source types.Source, openai_key string, database_url string) (types.ExpandedSource, bool) {
	// Since wikinews external links don't provide a publication date,
	// we use the current time and assume freshness.
	es := types.ExpandedSource{
		Title:  source.Title,
		Link:   source.Link,
		Date:   source.Date,
		Origin: source.Origin,
	}

	// Apply standard filters (skip freshness check since we assume fresh)
	filters_list := []types.Filter{
		filters.IsFreshFilter(),
		filters.IsDupeFilter(database_url),
		filters.CleanTitleFilter(),
		filters.ExtractBetterTitle(),
		filters.ExtractSummaryFilter(openai_key),
		filters.CheckImportanceFilter(openai_key),
	}
	es, ok := filters.ApplyFilters(es, filters_list)
	return es, ok

}
