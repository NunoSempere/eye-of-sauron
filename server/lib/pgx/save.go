package pgx

import (
	"context"
	"log"
	"os"

	"git.nunosempere.com/NunoSempere/news/lib/types"
	"github.com/jackc/pgx/v5"
)

// SaveToMainDatabase saves the source to the main sources table
func SaveToMainDatabase(source types.ExpandedSource) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_POOL_URL"))
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

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

// SaveToAIDatabase saves the source to the sources-ai table
func SaveToAIDatabase(source types.ExpandedSource) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_POOL_URL"))
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), `
        INSERT INTO "sources-ai" (title, link, date, summary, importance_bool, importance_reasoning)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (link) DO NOTHING
    `, source.Title, source.Link, source.Date, source.Summary, source.ImportanceBool, source.ImportanceReasoning)

	if err != nil {
		log.Printf("Error saving source to sources-ai table: %v\n", err)
		return
	}

	log.Printf("Saved source to sources-ai table: %v\n", source.Title)
}

// SaveSource saves to main database (standard behavior)
func SaveSource(source types.ExpandedSource) {
	SaveToMainDatabase(source)
}

// SaveSourceConditional saves to AI database always, and to main database if passes_filters is true
func SaveSourceConditional(source types.ExpandedSource, passes_filters bool) {
	SaveToAIDatabase(source)
	if passes_filters {
		SaveToMainDatabase(source)
	}
}
