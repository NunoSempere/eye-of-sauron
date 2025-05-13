package main

import (
	"encoding/json"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"log"
	"os"
	"path/filepath"
	"time"
)

func SaveSource(source types.ExpandedSource) {
	// Create the output directory if it doesn't exist
	outputDir := "output/potpourri/hn"
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Printf("Error creating output directory: %v", err)
		return
	}

	// Create a filename based on the current timestamp
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := filepath.Join(outputDir, timestamp+".json")

	// Marshal the source to JSON
	jsonData, err := json.MarshalIndent(source, "", "  ")
	if err != nil {
		log.Printf("Error marshaling source to JSON: %v", err)
		return
	}

	// Write the JSON to file
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		log.Printf("Error writing JSON to file: %v", err)
		return
	}

	log.Printf("Saved source to %s", filename)
}
