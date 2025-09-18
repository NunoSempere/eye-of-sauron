package main

import (
	"io"
	"log"
	"os"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/readability"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"github.com/joho/godotenv"
)

func main() {
	// Set up logging
	logFile, err := os.OpenFile("v2.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	// Load environment variables
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	openai_key := os.Getenv("OPENAI_KEY")
	pg_database_url := os.Getenv("DATABASE_POOL_URL")

	for {
		log.Println("Starting Wikipedia current events processing")
		rssURL := "https://www.to-rss.xyz/wikipedia/current_events/"

		link, err := ExtractCurrentEventsLink(rssURL)
		if err != nil {
			log.Printf("Error extracting current events link: %v", err)
			continue
		}
		if link == "" {
			log.Printf("No current events link found")
			continue
		}
		log.Printf("Current events link: %s", link)

		// Fetch the content
		content, err := FetchCurrentEvents(link)
		if err != nil {
			log.Printf("Error fetching current events: %v", err)
			continue
		}

		// Extract and process external links
		externalLinks := ExtractExternalLinks(content)
		log.Printf("Found %d external news source links", len(externalLinks))

		// Process each external link
		for i, extLink := range externalLinks {
			log.Printf("\nProcessing link %d/%d: %s", i+1, len(externalLinks), extLink)
			source := types.Source{
				Title: extLink,
				Link:  extLink,
				Date:  time.Now(),
				Origin: "Wikinews",
			}

			// Initialize expanded source
			es := types.ExpandedSource{
				Title: source.Title,
				Link:  source.Link,
				Date:  source.Date,
				Origin: source.Origin,
			}

			// Apply standard filters (skip freshness check since we assume fresh)
			filters_list := []types.Filter{
				filters.IsDupeFilter(pg_database_url),
				filters.IsGoodHostFilter(),
				filters.CleanTitleFilter(),
			}
			es, ok := filters.ApplyFilters(es, filters_list)
			if !ok {
				continue
			}

			// Try to get a better title from the source HTML
			if title := readability.ExtractTitle(source.Link); title != "" {
				es.Title = title
				log.Printf("Found title from HTML: %s", title)
				// Clean the extracted title
				es.Title = filters.CleanTitle(es.Title)
			}

			// Extract content and summarize
			es, ok = filters.ExtractContentAndSummarize(es, openai_key)
			if !ok {
				continue
			}

			// Check importance
			expanded_source, passes_filters := filters.CheckImportance(es, openai_key)
			if passes_filters {
				SaveSource(expanded_source)
			}
		}

		log.Printf("Finished processing current events, sleeping for 24 hours")
		time.Sleep(12 * time.Hour)
	}
}
