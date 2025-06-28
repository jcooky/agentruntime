package config

import (
	"fmt"
	"os"
)

type MemoryConfig struct {
	// Core Database Settings
	// SqliteEnabled controls whether SQLite memory service is activated
	// Default: true
	SqliteEnabled bool `json:"sqliteEnabled,omitempty"`

	// SqlitePath specifies the file path for the SQLite database
	// Default: ~/.agentruntime/memory.db
	SqlitePath string `json:"sqlitePath,omitempty"`

	// Vector Search Settings
	// VectorEnabled controls whether vector search functionality is enabled
	// Requires OpenAI API key for embeddings
	// Default: true
	VectorEnabled bool `json:"vectorEnabled,omitempty"`

	// Rerank Settings
	// RerankEnabled controls whether to use LLM-based reranking after vector search
	// This improves search accuracy by evaluating semantic relevance
	// Default: true
	RerankEnabled bool `json:"rerankEnabled,omitempty"`

	// RerankModel specifies which LLM model to use for reranking
	// Supports OpenAI models (e.g., "gpt-4o-mini", "gpt-4") or other providers
	// Default: "gpt-4o-mini"
	RerankModel string `json:"rerankModel,omitempty"`

	// RerankTopK sets the maximum number of results to return after reranking
	// This is the final number of results presented to the user
	// Default: 10
	RerankTopK int `json:"rerankTopK,omitempty"`

	// RetrievalFactor determines how many candidates to retrieve for reranking
	// Actual retrieval count = RerankTopK × RetrievalFactor
	// Higher values give reranker more options but increase latency
	// Default: 3 (retrieves 3x the final result count)
	RetrievalFactor int `json:"retrievalFactor,omitempty"`

	// UseBatchRerank controls whether to use batch processing for reranking
	// Batch processing evaluates all candidates in a single LLM call (more efficient)
	// Individual processing makes separate calls for each candidate (more accurate)
	// Default: true
	UseBatchRerank bool `json:"useBatchRerank,omitempty"`
}

// NewMemoryConfig creates a new MemoryConfig with sensible defaults
// These defaults can be overridden by environment variables
func NewMemoryConfig() *MemoryConfig {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return &MemoryConfig{
		// Core Database Settings
		SqliteEnabled: true,
		SqlitePath:    fmt.Sprintf("%s/.agentruntime/memory.db", home),

		// Vector Search Settings
		VectorEnabled: true,

		// Rerank Settings
		RerankEnabled:   true,
		RerankModel:     "gpt-4o-mini",
		RerankTopK:      10,
		RetrievalFactor: 3,    // Retrieve 3x candidates for reranking
		UseBatchRerank:  true, // Use batch reranker for better performance
	}
}
