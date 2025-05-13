package main

import (
	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/llm"
	"git.nunosempere.com/NunoSempere/news/lib/readability"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"log"
	"strings"
	"time"
)

func FilterAndExpandSource(source HNHit, openai_key string, database_url string) (types.ExpandedSource, bool) {
	// Parse the created_at time
	createdAt, err := time.Parse(time.RFC3339, source.CreatedAt)
	if err != nil {
		log.Printf("Could not parse date '%s', using current time", source.CreatedAt)
		createdAt = time.Now()
	}

	tmp_source := types.Source{
		Title:  source.Title,
		Link:   source.URL,
		Date:   createdAt.Format(time.RFC3339),
		Origin: "HackerNews",
	}
	expanded_source := types.ExpandedSource{
		Title:  source.Title,
		// Summary: source.StoryText, : will add later
		Link:   source.URL,
		Date:   createdAt.Format(time.RFC3339),
		Origin: "HackerNews",
	}

	if source.URL == "" && source.StoryText == "" {
		log.Printf("No url or text")
		return expanded_source, false
	}
	if source.Points < 2 && source.NumComments < 2 {
		log.Printf("< 2 points and < 2comments")
		return expanded_source, false
	}
	if startsWithAny(source.Title, []string{"Ask HN:", "Launch HN:", "Show HN:"}) {
		log.Printf("Ask/Launch/Show HN")
		return expanded_source, false
	}

	// Check for duplicates
	is_dupe := filters.IsDupe(tmp_source, database_url)
	if is_dupe {
		return expanded_source, false
	}

	// Check if host is acceptable
	is_good_host := filters.IsGoodHost(tmp_source)
	if !is_good_host {
		return expanded_source, false
	}

	// Clean the title
	expanded_source.Title = filters.CleanTitle(expanded_source.Title)

	// If story text is larger than a tweet, evaluate as is
	if len(source.StoryText) > 100 {
		expanded_source.Summary = source.StoryText
	} else {
		// Get article content
		content, err := readability.GetArticleContent(tmp_source.Link)
		if err != nil {
			log.Printf("Content extraction failed for %s: %v", tmp_source.Link, err)
			return expanded_source, false
		}
		
		// Summarize the article
		summary, err := llm.Summarize(content, openai_key)
		if err != nil {
			log.Printf("Summarization failed for %s: %v", tmp_source.Link, err)
			return expanded_source, false
		}
		expanded_source.Summary = summary
		log.Printf("Summary: %s", expanded_source.Summary)
	}

	// Check importance
	existential_importance_snippet := "# " + expanded_source.Title + "\n\n" + expanded_source.Summary
	existential_importance_box, err := llm.CheckExistentialImportance(existential_importance_snippet, openai_key)
	if err != nil || existential_importance_box == nil {
		log.Printf("Importance check failed for %s: %v", tmp_source.Link, err)
		return expanded_source, false
	}
	expanded_source.ImportanceBool = existential_importance_box.ExistentialImportanceBool
	expanded_source.ImportanceReasoning = existential_importance_box.ExistentialImportanceReasoning
	log.Printf("Importance bool: %t", expanded_source.ImportanceBool)
	log.Printf("Reasoning: %s", expanded_source.ImportanceReasoning)

	return expanded_source, expanded_source.ImportanceBool
}


func startsWithAny(s string, prefixes []string) bool {
        for _, prefix := range prefixes {
                if strings.HasPrefix(s, prefix) {
                        return true
                }
        }
        return false
}

