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

	// Custom content extraction for DSCA articles
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

	// Summarize the article
	summary, err := llm.Summarize(content, openai_key)
	if err != nil {
		log.Printf("Summarization failed for %s: %v", source.Link, err)
		return es, false
	}
	es.Summary = summary
	log.Printf("Summary: %s", es.Summary)

	// Check importance
	return filters.CheckImportance(es, openai_key)
}
