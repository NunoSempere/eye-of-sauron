package main

import (
	"context"
	"log"
	"os"

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

	// Insert into database with conflict handling
	_, err = conn.Exec(context.Background(), `
        INSERT INTO sources (title, link, date, summary, importance_bool, importance_reasoning)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (link) DO NOTHING
    `, source.Title, source.Link, source.Date, source.Summary, source.ImportanceBool, source.ImportanceReasoning)

	if err != nil {
		log.Printf("Error saving source to database: %v\n", err)
		return
	}

	log.Printf("Saved source: %v\n", source.Title)
}
