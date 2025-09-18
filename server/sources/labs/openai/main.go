package main

import (
	"io"
	"log"
	"os"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/llm"
	"git.nunosempere.com/NunoSempere/news/lib/readability"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"github.com/joho/godotenv"
)

func main() {
	// Set up logging
	logFile, err := os.OpenFile("v2.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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
		log.Println("Starting OpenAI news processing")
		
		sources, err := FetchSources()
		if err != nil {
			log.Printf("Error fetching sources: %v", err)
			continue
		}
		
		log.Printf("Found %d sources", len(sources))
		
		// Process each source
		for i, source := range sources {
			log.Printf("\nProcessing source %d/%d: %s", i+1, len(sources), source.Title)
			
			// Initialize expanded source with basic info
			expanded_source := types.ExpandedSource{
				Title: source.Title,
				Link:  source.Link,
				Date:  source.Date,
				Origin: "OpenAI",
			}

			// Check for duplicates
			if filters.IsDupe(source, pg_database_url) {
				log.Printf("Duplicate source found: %s", source.Link)
				continue
			}

			// Check if host is acceptable
			if !filters.IsGoodHost(source) {
				log.Printf("Host not acceptable: %s", source.Link)
				continue
			}

			// Try to get a better title from the source HTML
			if title := readability.ExtractTitle(source.Link); title != "" {
				expanded_source.Title = title
				log.Printf("Found title from HTML: %s", title)
			}

			// Clean up the title
			expanded_source.Title = filters.CleanTitle(expanded_source.Title)

			// Get article content using a readability extractor
			content, err := readability.GetArticleContent(source.Link)
			if err != nil {
				log.Printf("Readability extraction failed for %s: %v", source.Link, err)
				continue
			}

			// Summarize the article using an LLM
			summary, err := llm.Summarize(content, openai_key)
			if err != nil {
				log.Printf("Summarization failed for %s: %v", source.Link, err)
				continue
			}
			expanded_source.Summary = summary
			log.Printf("Summary: %s", expanded_source.Summary)

			// Check existential or importance threshold
			existential_importance_snippet := "# " + expanded_source.Title + "\n\n" + summary
			existential_importance_box, err := llm.CheckExistentialImportance(existential_importance_snippet, openai_key)
			if err != nil || existential_importance_box == nil {
				log.Printf("Importance check failed for %s: %v", source.Link, err)
				continue
			}
			expanded_source.ImportanceBool = existential_importance_box.ExistentialImportanceBool
			expanded_source.ImportanceReasoning = existential_importance_box.ExistentialImportanceReasoning
			log.Printf("Importance bool: %t", expanded_source.ImportanceBool)
			log.Printf("Reasoning: %s", expanded_source.ImportanceReasoning)

			// Always save to AI database, and save to main database if passes filters
			SaveSource(expanded_source, expanded_source.ImportanceBool)
		}

		log.Printf("Finished processing OpenAI news, sleeping for 6 hours")
		time.Sleep(6 * time.Hour)
	}
}
