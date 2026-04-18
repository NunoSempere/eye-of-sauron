package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

// Source represents an item from the sources table
type Source struct {
	ID                    int
	Title                 string
	Link                  string
	Date                  time.Time
	Summary               *string
	ImportanceBool        *bool
	ImportanceReasoning   *string
	CreatedAt             time.Time
	Processed             bool
	RelevantPerHumanCheck string
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	databaseURL := os.Getenv("DATABASE_POOL_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_POOL_URL environment variable not set")
	}

	// Connect to database
	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	// Query for items added in the last 24 hours
	query := `
		SELECT
			id,
			title,
			link,
			date,
			summary,
			importance_bool,
			importance_reasoning,
			created_at,
			processed,
			relevant_per_human_check
		FROM sources
		WHERE created_at >= NOW() - INTERVAL '24 hours'
		ORDER BY created_at DESC
	`

	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	// Fetch and display results
	var count int
	fmt.Println("Items added in the last 24 hours:")
	fmt.Println("=====================================")

	for rows.Next() {
		var source Source
		err := rows.Scan(
			&source.ID,
			&source.Title,
			&source.Link,
			&source.Date,
			&source.Summary,
			&source.ImportanceBool,
			&source.ImportanceReasoning,
			&source.CreatedAt,
			&source.Processed,
			&source.RelevantPerHumanCheck,
		)
		if err != nil {
			log.Printf("Error scanning row: %v\n", err)
			continue
		}

		count++
		fmt.Printf("\n[%d] %s\n", source.ID, source.Title)
		fmt.Printf("    Link: %s\n", source.Link)
		fmt.Printf("    Date: %s\n", source.Date.Format("2006-01-02 15:04:05"))
		fmt.Printf("    Created: %s\n", source.CreatedAt.Format("2006-01-02 15:04:05"))

		if source.Summary != nil {
			fmt.Printf("    Summary: %s\n", *source.Summary)
		}

		if source.ImportanceBool != nil {
			fmt.Printf("    Important: %v\n", *source.ImportanceBool)
			if source.ImportanceReasoning != nil {
				fmt.Printf("    Reasoning: %s\n", *source.ImportanceReasoning)
			}
		}

		fmt.Printf("    Processed: %v\n", source.Processed)
		fmt.Printf("    Human Check: %s\n", source.RelevantPerHumanCheck)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v\n", err)
	}

	fmt.Printf("\n=====================================\n")
	fmt.Printf("Total items found: %d\n", count)
}
