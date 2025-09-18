package main

import (
	"log"
	"os"
	"io"
	"time"

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
		log.Println("Starting OpenAI news processing")
		
		sources, err := FetchSources()
		if err != nil {
			log.Printf("Error fetching sources: %v", err)
			continue
		}
		
		log.Printf("Found %d sources", len(sources))
		
		// Process each source
		for i, source := range sources {
			log.Printf("\nProcessing source %d/%d: %s", i+1, len(sources), source.Title)
			
			expanded_source, passes_filters := FilterAndExpandSource(source, openai_key, pg_database_url)
			if passes_filters {
				SaveSource(expanded_source)
			}
		}

		log.Printf("Finished processing OpenAI news, sleeping for 6 hours")
		time.Sleep(6 * time.Hour)
	}
}
