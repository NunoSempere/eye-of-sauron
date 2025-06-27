package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

func main() {

	fmt.Println("Getting .env")
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v", err)
		return
	}
	openai_key := os.Getenv("OPENAI_KEY")

	fmt.Println("Getting titles")
	titles, err := ReadFileLines("data/titles.txt") // Replace with your filename
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Getting embeddings")
	embeddings, err := GetEmbeddings(titles, openai_key)
	if err != nil {
		fmt.Printf("Error getting embeddings: %v", err)
		return
	}

	fmt.Println("Calculating clusters")
	clusters := GetClusters(embeddings)
	for i, c := range clusters {
		fmt.Printf("\nCluster #%d\n", i)
		
		// Display centroid info
		if c.centroid != nil {
			fmt.Printf("  Centroid: [%.4f, %.4f, %.4f, ...]\n", c.centroid[0], c.centroid[1], c.centroid[2])
		}
		
		fmt.Printf("\n  Central points:\n")
		for _, j := range c.points {
			distance := calculateDistance(embeddings[j], c.centroid)
			fmt.Printf("    %s (distance: %.4f)\n", titles[j], distance)
		}
		fmt.Printf("\n  Related outliers:\n")
		for _, k := range c.outliers {
			distance := calculateDistance(embeddings[k], c.centroid)
			fmt.Printf("    %s (distance: %.4f)\n", titles[k], distance)
		}
	}

}
