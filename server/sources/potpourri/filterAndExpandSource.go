package main

import (
	"log"
	"strings"

	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/llm"
	"git.nunosempere.com/NunoSempere/news/lib/readability"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"git.nunosempere.com/NunoSempere/news/sources/potpourri/dsca"
)

func FilterAndExpandSource(source types.Source, openai_key string, database_url string) (types.ExpandedSource, bool) {
	// Initialize expanded source
	es := types.ExpandedSource{
		Title:  source.Title,
		Link:   source.Link,
		Date:   source.Date,
		Origin: source.Origin,
	}

	// Define custom filter
	TweakedSummaryFilter := func(source types.ExpandedSource) (types.ExpandedSource, bool) {
		var content string
		var err error

		if strings.Contains(source.Origin, "DSCA") {
			content, err = dsca.GetArticleContent(source.Link)
		} else {
			content, err = readability.GetArticleContent(source.Link)
		}
		if err != nil {
			log.Printf("Content extraction failed for %s: %v", source.Link, err)
			return es, false
		}
		summary, err := llm.Summarize(content, openai_key)
		if err != nil {
			log.Printf("Summarization failed for %s: %v", source.Link, err)
			return es, false
		}
		es.Summary = summary
		return es, true
	}

	fs := []types.Filter{
		filters.IsFreshFilter(),
		filters.IsDupeFilter(database_url),
		filters.IsGoodHostFilter(),
		filters.CleanTitleFilter(),
		TweakedSummaryFilter,
		filters.CheckImportanceFilter(openai_key),
	}
	es, ok := filters.ApplyFilters(es, fs)
	if !ok {
		return es, false
	}

	return es, ok
}
