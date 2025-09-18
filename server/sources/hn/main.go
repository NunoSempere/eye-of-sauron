package main

import (
	"io"
	"log"
	"os"
	"strings"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/filters"
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

	ticker_gkg := time.NewTicker(60 * time.Minute)
	defer ticker_gkg.Stop()
	for ; true; <-ticker_gkg.C {
		go func() {

			log.Println("Starting HackerNews processing")

			hnSources, err := FetchFeed()
			if err != nil {
				log.Printf("Error fetching HackerNews feed: %v", err)
			} else {
				log.Printf("Found %d HackerNews articles", len(hnSources))
				for i, hit := range hnSources {
					log.Printf("\nProcessing HackerNews article %d/%d: %s", i+1, len(hnSources), hit.Title)

					// Parse the created_at time
					createdAt, err := time.Parse(time.RFC3339, hit.CreatedAt)
					if err != nil {
						log.Printf("Could not parse date '%s', using current time", hit.CreatedAt)
						createdAt = time.Now()
					}

					// Initialize expanded source
					es := types.ExpandedSource{
						Title: hit.Title,
						Link:  hit.URL,
						Date:  createdAt,
						Origin: "HackerNews",
					}

					// HN-specific pre-filters
					if hit.URL == "" && hit.StoryText == "" {
						log.Printf("No url or text")
						continue
					}
					if hit.Points < 2 && hit.NumComments < 2 {
						log.Printf("< 2 points and < 2comments")
						continue
					}
					if startsWithAny(hit.Title, []string{"Ask HN:", "Launch HN:", "Show HN:"}) {
						log.Printf("Ask/Launch/Show HN")
						continue
					}

					// Apply standard filters
					filters_list := filters.StandardFilterPipeline(pg_database_url)
					es, ok := filters.ApplyFilters(es, filters_list)
					if !ok {
						continue
					}

					// Custom content handling for HN
					if len(hit.StoryText) > 100 {
						// Use story text directly if substantial
						es.Summary = hit.StoryText
					} else {
						// Extract and summarize content
						es, ok = filters.ExtractContentAndSummarize(es, openai_key)
						if !ok {
							continue
						}
					}

					// Check importance
					es, ok = filters.CheckImportance(es, openai_key)
					if !ok {
						continue
					}

					// HN-specific importance boost
					if strings.Contains(hit.Title, "Saudi Arabia") {
						es.ImportanceBool = true
						es.ImportanceReasoning = "Contains keyword"
					}

					if es.ImportanceBool {
						SaveSource(es)
					}
				}
			}
			log.Printf("Finished processing HackerNews, sleeping for 1 hour")
		}()
	}
}

func startsWithAny(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
