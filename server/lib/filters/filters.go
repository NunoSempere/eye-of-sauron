package filters

import (
	"context"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"github.com/jackc/pgx/v5"
	"log"
	"net/url"
	"slices"
	"strings"
	"time"
)

func IsFresh(source types.Source, layout string) bool {
	date_str := source.Date
	parsed_time, err := time.Parse(layout, date_str)
	if err != nil {
		log.Printf("Error parsing date: %v", err)
		return false
	}

	now := time.Now()
	fifteen_days_before := now.AddDate(0, 0, -15)
	fifteen_days_after := now.AddDate(0, 0, 15)

	return parsed_time.After(fifteen_days_before) && parsed_time.Before(fifteen_days_after)

}

func IsDupe(source types.Source, database_url string) bool {
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
	`, source.Title, source.Link).Scan(&exists)
	if err != nil {
		log.Printf("Error checking for duplicates: %v\n", err)
		return false
	}

	if exists {
		log.Printf("Skipping duplicate title/link: %v %v\n", source.Title, source.Link)
	} else {
		log.Printf("Article is not a duplicate")
	}
	return exists
}

func IsGoodHost(source types.Source) bool {
	parsedURL, err := url.Parse(source.Link)
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

func CleanTitle0(s string, endingMarker string) string {
	// endingMarkers: "-", "|"
	result := s
	if len(result) > 25 {
		// Logic seems good, but some titles are abridged somehow.
		if pos := strings.LastIndex(s[25:], endingMarker); pos != -1 {
			result = s[:25+pos]
		}
	}
	return result
}

func CleanTitle(s string) string {
	s2 := CleanTitle0(s, " – ")
	s3 := CleanTitle0(s2, " - ")
	s4 := CleanTitle0(s3, "|")
  s5 := strings.ReplaceAll(s4, "<b>", "")
	s6 := strings.ReplaceAll(s5, "</b>", "")
	s7 := strings.ReplaceAll(s6, "&#39;", "'")

	s8 := strings.TrimSpace(s7)
	return s8
}


/*
func main(){
	result := "25-day-old baby pulled from Gaza rubble after Israeli strike killed family"
	result2 := CleanTitle(result, "-")
	fmt.Println(result2)
}
*/
