package main

import (
	"context"
	"log"
	"os"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/types"
	"github.com/jackc/pgx/v5"
)

// SaveSource connects to the database and inserts the expanded source.
func SaveSource(source types.ExpandedSource) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_POOL_URL"))
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	// Parse the date - handle both RFC3339 and other common formats
	var date time.Time
	if source.Date != "" {
		// Try RFC3339 first
		date, err = time.Parse(time.RFC3339, source.Date)
		if err != nil {
			// Try RFC1123Z (common in RSS feeds)
			date, err = time.Parse(time.RFC1123Z, source.Date)
			if err != nil {
				// Try other common formats as needed
				log.Printf("Error parsing date %v: %v, using current time\n", source.Date, err)
				date = time.Now()
			}
		}
	} else {
		date = time.Now()
	}

	// Insert into database with conflict handling
	_, err = conn.Exec(context.Background(), `
        INSERT INTO sources (title, link, date, summary, importance_bool, importance_reasoning)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (link) DO NOTHING
    `, source.Title, source.Link, date, source.Summary, source.ImportanceBool, source.ImportanceReasoning)

	if err != nil {
		log.Printf("Error saving source to database: %v\n", err)
		return
	}
	
	log.Printf("Saved source: %v\n", source.Title)
}