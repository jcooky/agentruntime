package knowledge

import (
	"context"
	"time"
)

// Store defines the interface for complete knowledge storage operations
type Store interface {
	// Store stores or updates knowledge items with their embeddings
	// This should be atomic - either all items are stored or none
	Store(ctx context.Context, items []KnowledgeItem) error

	// Search performs semantic search and returns matching results
	Search(ctx context.Context, agentName string, queryEmbedding []float32, limit int) ([]KnowledgeSearchResult, error)

	// GetByAgent retrieves all knowledge items for a specific agent
	GetByAgent(ctx context.Context, agentName string) ([]KnowledgeItem, error)

	// DeleteByAgent removes all knowledge for a specific agent
	DeleteByAgent(ctx context.Context, agentName string) error

	// Close closes the knowledge store and releases resources
	Close() error
}

// KnowledgeItem represents a complete knowledge entry with all its data
type KnowledgeItem struct {
	ID        string                 // Unique identifier
	AgentName string                 // Agent this knowledge belongs to
	Content   string                 // Text content
	Embedding []float32              // Vector embedding
	Metadata  map[string]interface{} // Additional metadata
	CreatedAt time.Time              // Creation timestamp
	UpdatedAt time.Time              // Last update timestamp
}

// KnowledgeSearchResult represents a search result with similarity score
type KnowledgeSearchResult struct {
	KnowledgeItem
	Score float32 // Similarity score (0-1, higher is better)
}
