package main

import (
	"log"
	"regexp"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/llm"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

func ExtractDateFromURL(url string) (time.Time, bool) {
	// Pattern matches URLs like "https://mil.gmw.cn/2025-02/10/content_37841910.htm"
	pattern := regexp.MustCompile(`/(\d{4})-(\d{2})/(\d{2})/`)
	matches := pattern.FindStringSubmatch(url)

	if len(matches) == 4 {
		year := matches[1]
		month := matches[2]
		day := matches[3]
		dateStr := year + "-" + month + "-" + day
		date, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			return date, true
		}
	}
	return time.Now(), false
}

func IsWithinTwoDays(articleDate time.Time) bool {
	oneWeekAgo := time.Now().AddDate(0, 0, -2)
	return articleDate.After(oneWeekAgo)
}

func TranslateArticle(article GmwMilSource, openai_token string) (GmwMilSourceTranslated, error) {
	translated_title, err := llm.TranslateString(article.Title, openai_token)
	if err != nil {
		return GmwMilSourceTranslated{}, err
	}
	translated_content, err := llm.TranslateString(article.Content, openai_token)
	if err != nil {
		return GmwMilSourceTranslated{}, err
	}
	return GmwMilSourceTranslated{
		Link:            article.Link,
		OriginalTitle:   article.Title,
		OriginalContent: article.Content,
		EnglishTitle:    translated_title,
		EnglishContent:  translated_content,
	}, nil
}

func FilterAndExpandSource(article GmwMilSource, openai_key string, database_url string) (types.ExpandedSource, bool) {
	// is there some parsimonious way of doing this in a more functional way that reuses code? Probably not?

	date, _ := ExtractDateFromURL(article.Link)
	is_dupe := filters.IsDupe(types.Source{Title: article.Title, Link: article.Link}, database_url)
	if is_dupe {
		return types.ExpandedSource{}, false
	}

	gmw, err := TranslateArticle(article, openai_key)
	if err != nil {
		log.Printf("%v", err)
		return types.ExpandedSource{}, false
	}
	log.Printf("\nTranslated title: %s", gmw.EnglishTitle)

	expanded_source := types.ExpandedSource{
		Title: gmw.EnglishTitle,
		Link:  gmw.Link,
		Date:  date,
	}

	summary, err := llm.Summarize(gmw.EnglishContent+"\n\nWhen summarizing a Chinese article, give the gist in idiomatic English, rather than selecting the most important phrases in Chinese", openai_key)
	if err != nil {
		log.Printf("%v", err)
		return expanded_source, false
	}
	expanded_source.Summary = summary
	log.Printf("\nSummary: %s", expanded_source.Summary)

	existential_importance_snippet := "# " + expanded_source.Title + "\n\n" + summary
	existential_importance_box, err := llm.CheckExistentialImportanceChina(existential_importance_snippet, openai_key)
	if err != nil || existential_importance_box == nil {
		log.Printf("%v", err)
		return expanded_source, false
	}
	expanded_source.ImportanceBool = existential_importance_box.ExistentialImportanceBool
	expanded_source.ImportanceReasoning = existential_importance_box.ExistentialImportanceReasoning

	log.Printf("Importance bool: %t", expanded_source.ImportanceBool)
	log.Printf("Importance reasoning: %s", expanded_source.ImportanceReasoning)

	return expanded_source, expanded_source.ImportanceBool
}
