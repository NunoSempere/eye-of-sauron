package main

import (
	"io"
	"log"
	"os"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/pgx"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"github.com/joho/godotenv"
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

				es := types.ExpandedSource{Title: article.Title, Link: article.Link, Date: article.Date}

				fs := []types.Filter{
					filters.IsFreshFilter(),
					filters.IsDupeFilter(pg_database_url),
					filters.IsGoodHostFilter(),
					filters.CleanTitleFilter(),
					filters.ExtractSummaryFilter(openai_key),
					filters.CheckImportanceFilter(openai_key),
				}
				es, ok := filters.ApplyFilters(es, fs)
				if ok {
					pgx.SaveSource(es)
				}
			}
			log.Printf("\n\nFinished processing gkg batch\n")
		}()
	}

}
