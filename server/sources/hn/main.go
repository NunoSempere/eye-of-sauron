package main

import (
	"log"
	"os"
	"io"
	"time"

	"github.com/joho/godotenv"
	// "git.nunosempere.com/NunoSempere/news/lib/types"
)

func main() {
	// Set up logging
	logFile, err := os.OpenFile("sources/hn/log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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
		log.Println("Starting HackerNews processing")
		
		hnSources, err := FetchFeed()
		if err != nil {
			log.Printf("Error fetching HackerNews feed: %v", err)
		} else {
			log.Printf("Found %d HackerNews articles", len(hnSources))
			for i, source := range hnSources {
				log.Printf("\nProcessing HackerNews article %d/%d: %s", i+1, len(hnSources), source.Title)
				expanded_source, passes_filters := FilterAndExpandSource(source, openai_key, pg_database_url)
				if passes_filters {
					expanded_source.Origin = source.Origin
					SaveSource(expanded_source)
				}
			}
		}
		
		log.Printf("Finished processing HackerNews, sleeping for 1 hour")
		time.Sleep(1 * time.Hour)
	}
}
