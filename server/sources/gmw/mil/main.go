package main

import (
	"io"
	"log"
	"os"
	"math/rand"
	"slices"
	"time"

	"github.com/joho/godotenv"
)

func main() {

	// Initialize logging
	logFile, err := os.OpenFile("mil/v2.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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

	for {
	frontpage_articles, err := GetFrontpageUrls()
	if err != nil {
		log.Print(err)
	}
	// log.Println(frontpage_articles)
	titles := []string{}
	for _, url := range frontpage_articles {
		log.Printf("Url: %s", url)

		// filter here so as to not fetch full article if not necessary
		date, hasDate := ExtractDateFromURL(url)
		if hasDate && !IsWithinTwoDays(date) {
			log.Printf("Article is stale")
			continue
		}

		ms := 5000 + int64(2000 * rand.Float32())
		time.Sleep(time.Duration(ms) * time.Millisecond)
		article, err := ExtractFrontpageArticle(url)
		if err != nil {
			log.Print(err)
			// log.Print(err)
			continue
		}

		// All articles are duplicated, but with different underlying urls :(
		if slices.Contains(titles, article.Title) {
			// log.Println("Article is a duplicate")
			continue
		}
		titles = append(titles, article.Title)

		log.Printf("Title: %s", article.Title)

		expanded_source, passes_filters := FilterAndExpandSource(article, openai_key, pg_database_url)
		// useful for filtering duplicates within the same batch
		// eventually I'd need to have a longer database of titles
		if passes_filters {
			log.Println(expanded_source.Summary)
			SaveSource(expanded_source)
		}
	}
	log.Printf("Finished batch. Continuing in 12 hours")
	time.Sleep(12 * time.Hour)
	}
}
