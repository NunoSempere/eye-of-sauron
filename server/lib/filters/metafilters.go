package filters

import (
	"context"
	"log"
	"net/url"
	"slices"

	"git.nunosempere.com/NunoSempere/news/lib/llm"
	"git.nunosempere.com/NunoSempere/news/lib/readability"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"github.com/jackc/pgx/v5"

	// "strings"
	"time"
)

func isFreshTime(t time.Time) bool {
	now := time.Now()
	fifteen_days_before := now.AddDate(0, 0, -15)
	fifteen_days_after := now.AddDate(0, 0, 15)
	return t.After(fifteen_days_before) && t.Before(fifteen_days_after)
}

func IsFreshFilter() types.Filter {
	filter := func(source types.ExpandedSource) (types.ExpandedSource, bool) {
		return source, isFreshTime(source.Date)
	}
	return filter
}

func isDupeTitleOrLink(database_url string, title string, link string) bool {
	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		return false
	}
	defer conn.Close(context.Background())

	var exists bool
	err = conn.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM sources 
			WHERE UPPER(title) = $1 OR link = $2
		)
	`, title, link).Scan(&exists)
	if err != nil {
		log.Printf("Error checking for duplicates: %v\n", err)
		return false
	}

	if exists {
		log.Printf("Skipping duplicate title/link: %v %v\n", title, link)
	} else {
		log.Printf("Article is not a duplicate")
	}
	return exists
}

func IsDupeFilter(database_url string) types.Filter {
	filter := func(source types.ExpandedSource) (types.ExpandedSource, bool) {
		return source, !isDupeTitleOrLink(database_url, source.Title, source.Link)
	}
	return filter
}

func isGoodHost(link string) bool {
	parsedURL, err := url.Parse(link)
	if err != nil {
		log.Printf("Error parsing link: %v", err)
		return false
	}
	skippable_hosts := []string{"www.washingtonpost.com", "www.youtube.com", "www.naturalnews.com", "facebook.com", "m.facebook.com", "www.bignewsnetwork.com"}
	is_bad_host := slices.Contains(skippable_hosts, parsedURL.Host)
	if is_bad_host {
		log.Printf("Article is from a bad host")
	} else {
		log.Printf("Article is from a good host")
	}

	return !is_bad_host
}

func IsGoodHostFilter() types.Filter {
	filter := func(source types.ExpandedSource) (types.ExpandedSource, bool) {
		return source, isGoodHost(source.Link)
	}
	return filter
}

func CleanTitleFilter() types.Filter {
	filter := func(source types.ExpandedSource) (types.ExpandedSource, bool) {
		source.Title = CleanTitle(source.Title)
		return source, true
	}
	return filter

}

func ExtractContentAndSummarizeFilter(openai_key string) types.Filter {
	filter := func(source types.ExpandedSource) (types.ExpandedSource, bool) {
		content, err := readability.GetArticleContent(source.Link)
		if err != nil {
			log.Printf("Filtered because: Error getting article content: %v", err)
			return source, false
		}

		summary, err := llm.Summarize(content, openai_key)
		if err != nil {
			log.Printf("Filtered because: Error summarizing: %v", err)
			return source, false
		}
		source.Summary = summary
		return source, true
	}
	return filter
}

func CheckImportanceFilter(openai_key string) types.Filter {
	filter := func(source types.ExpandedSource) (types.ExpandedSource, bool) {
		existential_importance_snippet := "# " + source.Title + "\n\n" + source.Summary
		existential_importance_box, err := llm.CheckExistentialImportance(existential_importance_snippet, openai_key)
		if err != nil || existential_importance_box == nil {
			log.Printf("Filtered because: is not important")
			return source, false
		}
		source.ImportanceBool = existential_importance_box.ExistentialImportanceBool
		source.ImportanceReasoning = existential_importance_box.ExistentialImportanceReasoning

		log.Printf("importance bool: %t", source.ImportanceBool)
		return source, source.ImportanceBool
	}
	return filter
}
