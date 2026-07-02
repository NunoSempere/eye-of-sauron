package main

import(
	"context"
	"os"
	openrouter "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := openrouter.New(
        openrouter.WithSecurity(os.Getenv("OPENROUTER_API_KEY")),
    )

    res, err := s.Embeddings.Generate(ctx, operations.CreateEmbeddingsRequest{
        Input: operations.CreateInputUnionStr(
            "The quick brown fox jumps over the lazy dog",
        ),
        Model: "openai/text-embedding-3-small",
    })
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}

