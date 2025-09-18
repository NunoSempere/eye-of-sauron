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
		log.Println("Starting {{SOURCE_NAME}} processing")

		// TODO: Replace with your source-specific fetching logic
		sources, err := FetchSources()
		if err != nil {
			log.Printf("Error fetching sources: %v", err)
			continue
		}

		log.Printf("Found %d sources", len(sources))

		// Process each source
		for i, source := range sources {
			log.Printf("\nProcessing source %d/%d: %s", i+1, len(sources), source.Title)

			// Use standard processing pipeline with optional title extraction
			es := types.ExpandedSource{
				Title: source.Title,
				Link:  source.Link,
				Date:  source.Date,
				Origin: source.Origin,
			}

			// Apply standard filters
			filters_list := filters.StandardFilterPipeline(pg_database_url)
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

		log.Printf("Finished processing {{SOURCE_NAME}}, sleeping for {{SLEEP_DURATION}}")
		time.Sleep(10000) // TODO: Replace with appropriate duration
	}
}
