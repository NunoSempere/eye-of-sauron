package main

import (
	"github.com/joho/godotenv"
	"io"
	"log"
	"os"
	"time"
)

func main() {

	// Initialize logging
	logFile, err := os.OpenFile("v2.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	// Get keys
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	openai_key := os.Getenv("OPENAI_KEY")
	pg_database_url := os.Getenv("DATABASE_POOL_URL")

	// Search gkg
	ticker_gkg := time.NewTicker(15 * time.Minute)
	defer ticker_gkg.Stop()
	for ; true; <-ticker_gkg.C {
		go func() {
			// The prospector can be processing more than 2 GKG 15 minute intervals at the same time!
			log.Println("Processing new gkg batch (this may take a min or two, as it's a large zip file)")
			articles, err := SearchGKG()
			if err != nil {
				for i := 0; i < 2; i++ {
					log.Printf("GDELT.GKG error: %v", err)
					if i != 9 {
						log.Printf("trying again in 30s")
					}
					time.Sleep(30 * time.Second)
					articles, err = SearchGKG()
					if err == nil {
						break
					}
				}
				if err != nil {
					log.Printf("GDELT.GKG error: %v", err)
					log.Printf("Tried 10 times and couldn't parse GKG zip file")
					return

				}
			}
			log.Printf("Batch has %d articles\n", len(articles))
			for i, article := range articles {
				log.Printf("\n\nArticle #%v/%v [GDELT.GKG]: %v (%v)\n", i+1, len(articles), article.Title, article.Date)

				expanded_source, passes_filters := FilterAndExpandSource(article, openai_key, pg_database_url)
				if passes_filters {
					SaveSource(expanded_source)
				}
			}
			log.Printf("\n\nFinished processing gkg batch\n")
			return
		}()
	}

	// Keep main function alive
	/*
		done := make(chan bool)
		log.Println("Main function is now waiting indefinitely...")
		<-done
	*/

}
