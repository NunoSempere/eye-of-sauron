package main

import (
	"log"
	"os"
	"io"
	"time"

	"github.com/joho/godotenv"
	"git.nunosempere.com/NunoSempere/news/sources/potpourri/config"
	"git.nunosempere.com/NunoSempere/news/sources/potpourri/cnn"
	"git.nunosempere.com/NunoSempere/news/sources/potpourri/dsca"
	"git.nunosempere.com/NunoSempere/news/sources/potpourri/whitehouse"
)

func main() {
	// Set up logging
	logFile, err := os.OpenFile("sources/potpourri/v2.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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
						expanded_source, passes_filters := FilterAndExpandSource(source, openai_key, pg_database_url)
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
						expanded_source, passes_filters := FilterAndExpandSource(source, openai_key, pg_database_url)
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
						expanded_source, passes_filters := FilterAndExpandSource(source, openai_key, pg_database_url)
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
