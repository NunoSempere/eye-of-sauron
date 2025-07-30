package main

import (
	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/llm"
	"git.nunosempere.com/NunoSempere/news/lib/readability"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"log"
)

// Filters

func FilterAndExpandSource(source types.Source, openai_key string, database_url string) (types.ExpandedSource, bool) {
	expanded_source := types.ExpandedSource{Title: source.Title, Link: source.Link, Date: source.Date}

	is_dupe := filters.IsDupe(source, database_url)
	if is_dupe {
		log.Printf("Filtered because: is dupe")
		return expanded_source, false
	}

	is_fresh := filters.IsFresh(source, "2006-01-02T15:04:05Z")
	if !is_fresh {
		log.Printf("Filtered because: is old")
		return expanded_source, false
	}

	is_good_host := filters.IsGoodHost(source)
	if !is_good_host {
		log.Printf("Filtered because: is bad host")
		return expanded_source, false
	}

	expanded_source.Title = filters.CleanTitle(expanded_source.Title)

	content, err := readability.GetArticleContent(source.Link)
	if err != nil {
		log.Printf("Filtered because: Error getting article content: %v", err)
		return expanded_source, false
	}
	summary, err := llm.Summarize(content, openai_key)
	if err != nil {
		log.Printf("Filtered because: Error summarizing: %v", err)
		return expanded_source, false
	}
	expanded_source.Summary = summary

	existential_importance_snippet := "# " + source.Title + "\n\n" + summary
	existential_importance_box, err := llm.CheckExistentialImportance(existential_importance_snippet, openai_key)
	if err != nil || existential_importance_box == nil {
		log.Printf("Filtered because: is not important")
		return expanded_source, false
	}
	expanded_source.ImportanceBool = existential_importance_box.ExistentialImportanceBool
	expanded_source.ImportanceReasoning = existential_importance_box.ExistentialImportanceReasoning

	log.Printf("Importance bool: %b", expanded_source.ImportanceBool)

	return expanded_source, expanded_source.ImportanceBool
}
