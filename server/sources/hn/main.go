package main

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	// "git.nunosempere.com/NunoSempere/news/lib/types"
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
					expanded_source, passes_filters := FilterAndExpandSource(hit, openai_key, pg_database_url)
					if passes_filters {
						SaveSource(expanded_source)
					}
				}
			}
			log.Printf("Finished processing HackerNews, sleeping for 1 hour")
		}()
	}
}
