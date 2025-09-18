package main

import (
	"time"
	"log"

	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/llm"
	"git.nunosempere.com/NunoSempere/news/lib/readability"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

// FilterAndExpandSource processes an Anthropic news source through various filters,
// expands its content (via summarization and importance check),
// and returns an ExpandedSource and a boolean indicating if it passes thresholds.
func FilterAndExpandSource(source types.Source, openai_key string, database_url string) (types.ExpandedSource, bool) {
	// Initialize expanded source with basic info
	expanded_source := types.ExpandedSource{
		Title: source.Title,
		Link:  source.Link,
		Date:  source.Date,
	}

	// If no date provided, use current time
	if expanded_source.Date == "" {
		expanded_source.Date = time.Now().Format(time.RFC3339)
	}

	// Check for duplicates
	is_dupe := filters.IsDupe(source, database_url)
	if is_dupe {
		log.Printf("Duplicate source found: %s", source.Link)
		return expanded_source, false
	}

	// Anthropic news articles are generally fresh, no additional freshness check needed
	
	// Check if host is acceptable
	is_good_host := filters.IsGoodHost(source)
	if !is_good_host {
		log.Printf("Host not acceptable: %s", source.Link)
		return expanded_source, false
	}

	// Try to get a better title from the source HTML
	/*
	if title := readability.ExtractTitle(source.Link); title != "" {
		expanded_source.Title = title
		log.Printf("Found title from HTML: %s", title)
	}
	*/

	// Clean up the title
	expanded_source.Title = filters.CleanTitle(expanded_source.Title)

	// Get article content using a readability extractor
	content, err := readability.GetArticleContent(source.Link)
	if err != nil {
		log.Printf("Readability extraction failed for %s: %v", source.Link, err)
		return expanded_source, false
	}
	
	// Summarize the article using an LLM
	summary, err := llm.Summarize(content, openai_key)
	if err != nil {
		log.Printf("Summarization failed for %s: %v", source.Link, err)
		return expanded_source, false
	}
	expanded_source.Summary = summary
	log.Printf("Summary: %s", expanded_source.Summary)

	// Check existential or importance threshold
	existential_importance_snippet := "# " + expanded_source.Title + "\n\n" + summary
	existential_importance_box, err := llm.CheckExistentialImportance(existential_importance_snippet, openai_key)
	if err != nil || existential_importance_box == nil {
		log.Printf("Importance check failed for %s: %v", source.Link, err)
		return expanded_source, false
	}
	expanded_source.ImportanceBool = existential_importance_box.ExistentialImportanceBool
	expanded_source.ImportanceReasoning = existential_importance_box.ExistentialImportanceReasoning
	log.Printf("Importance bool: %t", expanded_source.ImportanceBool)
	log.Printf("Reasoning: %s", expanded_source.ImportanceReasoning)

	return expanded_source, expanded_source.ImportanceBool
}
