package main

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {

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

	keywords := []string{"War", "Emergency", "disaster", "alert", "nuclear", "combat duty", "human-to-human", "pandemic", "blockade", "invasion", "undersea cables", "nuclear", "Carrington event", "mystery pneumonia", "Taiwan", "Ukraine", "OpenAI announces AGI", "AI rights", "military exercise", "Kessler syndrome", "Cyberattack"}
	for {
		log.Println("(Re)starting Google Alerts keyword loop")
		for _, keyword := range keywords {
			log.Printf("Keyword: %v", keyword)
			articles, err := SearchGoogleAlerts(keyword)
			if err != nil {
				log.Printf("Google Alerts error: %v", err)
				continue
			}

			log.Printf("Number of articles in keyword: %v", len(articles))

			for i, article := range articles {
				log.Printf("\n")
				log.Printf("Article #%v/%v [keyword \"%v\"]: %v (%v)", i, len(articles), keyword, article.Title, article.Date)
				expanded_source, passes_filters := FilterAndExpandSource(article, openai_key, pg_database_url)
				if passes_filters {
					log.Printf("Saving source")
					SaveSource(expanded_source)
				} else {
					log.Printf("Not saving source")
				}
			}
		}
		log.Printf("Finished Google Alerts batch, pausing for half an hour")
		time.Sleep(1800 * time.Second) // stagger a little bit
	}

}
