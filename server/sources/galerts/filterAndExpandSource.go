package main

import (
	"log"

	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/llm"
	"git.nunosempere.com/NunoSempere/news/lib/readability"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

// Filters

func FilterAndExpandSource(source types.Source, openai_key string, database_url string) (types.ExpandedSource, bool) {
	fs := []types.Filter{
		filters.IsFreshFilter(),
		filters.IsDupeFilter(database_url),
		filters.IsGoodHostFilter(),
		filters.CleanTitleFilter(),
	}

	b := true
	es := types.ExpandedSource{Title: source.Title, Link: source.Link, Date: source.Date}
	for _, f := range fs {
		es, b = f(es)
		if !b {
			return es, b
		}
	}

	content, err := readability.GetArticleContent(source.Link)
	if err != nil {
		log.Printf("Filtered because: Error getting article content: %v", err)
		return es, false
	}
	summary, err := llm.Summarize(content, openai_key)
	if err != nil {
		log.Printf("Filtered because: Error summarizing: %v", err)
		return es, false
	}
	es.Summary = summary

	existential_importance_snippet := "# " + source.Title + "\n\n" + summary
	existential_importance_box, err := llm.CheckExistentialImportance(existential_importance_snippet, openai_key)
	if err != nil || existential_importance_box == nil {
		log.Printf("Filtered because: is not important")
		return es, false
	}
	es.ImportanceBool = existential_importance_box.ExistentialImportanceBool
	es.ImportanceReasoning = existential_importance_box.ExistentialImportanceReasoning

	log.Printf("importance bool: %t", es.ImportanceBool)

	return es, es.ImportanceBool
}
