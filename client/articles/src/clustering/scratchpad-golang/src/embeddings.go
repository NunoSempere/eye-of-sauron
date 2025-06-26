package main

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
)

func getEmbedding(s string, token string) ([]float64, error) {
	client := openai.NewClient(token)

	// Create an EmbeddingRequest for the user query
	queryReq := openai.EmbeddingRequest{
		Input: []string{s},
		Model: openai.AdaEmbeddingV2,
	}

	// Create an embedding for the user query
	queryResponse, err := client.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		return []float64{}, err
	}
	queryEmbedding := queryResponse.Data[0]
	f32s := queryEmbedding.Embedding
	f64s := make([]float64, len(f32s))
	// Convert each float32 to float64
	for i, v := range f32s {
		f64s[i] = float64(v)
	}
	return f64s, nil
}

func GetEmbeddings(texts []string, token string) ([][]float64, error) {
	es := [][]float64{}
	for _, t := range texts {
		e, err := getEmbedding(t, token)
		if err != nil {
			return [][]float64{}, err
		}
		fmt.Printf("Embedding for %s is [%f, %f, %f, ...]\n", t, e[0], e[1], e[2])
		es = append(es, e)
	}
	return es, nil
}
