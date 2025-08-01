package filters

import (
	"log"
	"context"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"github.com/jackc/pgx/v5"
	"net/url"
	"slices"
	// "strings"
	"time"
)

func isFreshTime(t time.Time) bool { 
	now := time.Now()
	fifteen_days_before := now.AddDate(0, 0, -15)
	fifteen_days_after := now.AddDate(0, 0, 15)
  return t.After(fifteen_days_before) && t.Before(fifteen_days_after)
}

func IsFreshFilter(layout string) types.Filter {
	filter := func(source types.ExpandedSource) (types.ExpandedSource, bool) {
		date_str := source.Date
		parsed_time, err := time.Parse(layout, date_str)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			return types.ExpandedSource{}, false
		}
		return source, isFreshTime(parsed_time)
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
  filter := func (source types.ExpandedSource) (types.ExpandedSource, bool) {
  	return source, isDupeTitleOrLink(database_url, source.Title, source.Link)
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
  filter := func (source types.ExpandedSource) (types.ExpandedSource, bool) {
  	return source, isGoodHost(source.Link)
	}
	return filter
} 

func CleanTitleFilter() types.Filter {
  filter := func (source types.ExpandedSource) (types.ExpandedSource, bool) {
  	source.Title = CleanTitle(source.Title)
  	return source, true
  }
	return filter
	
}
