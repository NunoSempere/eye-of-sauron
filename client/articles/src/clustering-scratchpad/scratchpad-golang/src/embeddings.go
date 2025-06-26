package main

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
)

func GetEmbeddings(texts []string, token string) ([][]float64, error) {
	
	client := openai.NewClient(token)

	// Create an EmbeddingRequest for the user query
	queryReq := openai.EmbeddingRequest{
		Input: texts,
		Model: openai.SmallEmbedding3,
	}

	// Create an embedding for the user query
	queryResponse, err := client.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		return [][]float64{}, err
	}

	es := [][]float64{}
	for i, e := range queryResponse.Data {
		f32s := e.Embedding
		f64s := make([]float64, len(f32s))
		for i, v := range f32s {
			f64s[i] = float64(v)
		}
		es = append(es, f64s)
		fmt.Printf("%s => [%f, %f, %f, ...]\n", texts[i], f64s[0], f64s[1], f64s[2])
	}

	return es, nil
}
