package memory

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type (
	// Embedder interface for generating embeddings
	Embedder interface {
		Embed(ctx context.Context, texts ...string) ([][]float32, error)
	}

	// GenkitEmbedder implements Embedder using genkit functionality
	GenkitEmbedder struct {
		genkit *genkit.Genkit
	}
)

// NewGenkitEmbedder creates a new embedder using genkit
func NewGenkitEmbedder(genkit *genkit.Genkit) Embedder {
	return &GenkitEmbedder{genkit: genkit}
}

// Embed generates embeddings for the given texts
func (e *GenkitEmbedder) Embed(ctx context.Context, texts ...string) ([][]float32, error) {
	embedder := genkit.LookupEmbedder(e.genkit, "openai", "text-embedding-3-small")

	resp, err := ai.Embed(ctx, embedder, ai.WithTextDocs(texts...))
	if err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(resp.Embeddings))
	for i, embedding := range resp.Embeddings {
		embeddings[i] = embedding.Embedding
	}

	return embeddings, nil
}
