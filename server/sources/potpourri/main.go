package main

import (
	"io"
	"log"
	"os"

	"time"

	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/llm"
	"git.nunosempere.com/NunoSempere/news/lib/readability"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"git.nunosempere.com/NunoSempere/news/sources/potpourri/config"
	"git.nunosempere.com/NunoSempere/news/sources/potpourri/cnn"
	"git.nunosempere.com/NunoSempere/news/sources/potpourri/dsca"
	"git.nunosempere.com/NunoSempere/news/sources/potpourri/whitehouse"
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
		log.Println("Starting potpourri sources processing")
		
		// Process sources in configured order
		for _, sourceCfg := range config.GetEnabledSources() {
			switch sourceCfg.Name {
			case "DSCA":
				log.Println("Processing DSCA feed")
				dscaSources, err := dsca.FetchFeed()
				if err != nil {
					log.Printf("Error fetching DSCA feed: %v", err)
				} else {
					log.Printf("Found %d DSCA articles", len(dscaSources))
					for i, source := range dscaSources {
						log.Printf("\nProcessing DSCA article %d/%d: %s", i+1, len(dscaSources), source.Title)
						expanded_source, passes_filters := processSourceWithCustomContent(source, openai_key, pg_database_url, true)
						if passes_filters {
							expanded_source.Origin = source.Origin
							SaveSource(expanded_source)
						}
					}
				}

			case "WhiteHouse":
				log.Println("Processing White House feed")
				whSources, err := whitehouse.FetchFeed()
				if err != nil {
					log.Printf("Error fetching White House feed: %v", err)
				} else {
					log.Printf("Found %d White House articles", len(whSources))
					for i, source := range whSources {
						log.Printf("\nProcessing White House article %d/%d: %s", i+1, len(whSources), source.Title)
						expanded_source, passes_filters := processSourceWithCustomContent(source, openai_key, pg_database_url, false)
						if passes_filters {
							expanded_source.Origin = source.Origin
							SaveSource(expanded_source)
						}
					}
				}

			case "CNN":
				log.Println("Processing CNN feeds")
				cnnSources, err := cnn.FetchAllFeeds()
				if err != nil {
					log.Printf("Error fetching CNN feeds: %v", err)
				} else {
					log.Printf("Found %d CNN articles", len(cnnSources))
					for i, source := range cnnSources {
						log.Printf("\nProcessing CNN article %d/%d [%s]: %s", i+1, len(cnnSources), source.Origin, source.Title)
						expanded_source, passes_filters := processSourceWithCustomContent(source, openai_key, pg_database_url, false)
						if passes_filters {
							expanded_source.Origin = source.Origin
							SaveSource(expanded_source)
						}
					}
				}

			}
		}
		
		log.Printf("Finished processing potpourri sources, sleeping for 1 hour")
		time.Sleep(1 * time.Hour)
	}
}

// processSourceWithCustomContent processes a source with custom content extraction logic
func processSourceWithCustomContent(source types.Source, openai_key string, database_url string, isDSCA bool) (types.ExpandedSource, bool) {
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
	if isDSCA {
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
