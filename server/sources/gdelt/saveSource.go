package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/types"
	"github.com/jackc/pgx/v5"
)

func SaveSource(source types.ExpandedSource) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_POOL_URL"))
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	layout := "20060102150405"
	date, err := time.Parse(layout, source.Date)
	if err != nil {
		log.Printf("Error parsing date %v in saveSource: %v\n", source.Date, err)
		return
	}

	_, err = conn.Exec(context.Background(), `
        INSERT INTO sources (title, link, date, summary, importance_bool, importance_reasoning)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (link) DO NOTHING
    `, source.Title, source.Link, date, source.Summary, source.ImportanceBool, source.ImportanceReasoning)

	if err != nil {
		log.Printf("Error saving source to database: %v\n", err)
		return
	}

	fmt.Printf("Saved source: %v", source.Title)
}
