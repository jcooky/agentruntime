package knowledge

import (
	"context"
)

// Store defines the interface for complete knowledge storage operations
type Store interface {
	// Store stores or updates knowledge items with their embeddings
	// This should be atomic - either all items are stored or none
	Store(ctx context.Context, knowledge *Knowledge) error

	// Search performs semantic search and returns matching results
	Search(ctx context.Context, queryEmbedding []float32, limit int, allowedKnowledgeIds []string) ([]KnowledgeSearchResult, error)

	// GetKnowledgeById retrieves all knowledge items for a specific agent
	GetKnowledgeById(ctx context.Context, knowledgeId string) (*Knowledge, error)

	// DeleteKnowledgeById removes all knowledge for a specific agent
	DeleteKnowledgeById(ctx context.Context, knowledgeId string) error

	// Close closes the knowledge store and releases resources
	Close() error
}
