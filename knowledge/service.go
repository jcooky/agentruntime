package knowledge

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/config"
	"github.com/pkg/errors"
)

type (
	Service interface {
		// Knowledge management methods
		IndexKnowledge(ctx context.Context, agentName string, knowledge []map[string]any) error
		RetrieveRelevantKnowledge(ctx context.Context, agentName string, query string, limit int) ([]string, error)
		DeleteAgentKnowledge(ctx context.Context, agentName string) error
		Close() error
	}

	service struct {
		store    Store
		embedder Embedder
		reranker Reranker
		config   *config.KnowledgeConfig
	}

	KnowledgeChunk struct {
		Content  string
		Metadata map[string]any
	}
)

var (
	_ Service = (*service)(nil)
)

// NewService creates a new knowledge service with default SQLite-based storage
func NewService(ctx context.Context, conf *config.KnowledgeConfig, logger *slog.Logger, genkit *genkit.Genkit) (Service, error) {
	if !conf.SqliteEnabled {
		return nil, errors.New("sqlite knowledge service is not enabled. Please check your configuration.")
	}
	if conf.SqlitePath == "" {
		return nil, errors.New("sqlite knowledge service path is not configured. Please check your configuration.")
	}

	// Create default SQLite knowledge store
	store, err := NewSqliteStore(conf.SqlitePath, 1536) // Default to OpenAI embedding dimension
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create SQLite knowledge store")
	}

	// Create embedder for RAG functionality
	embedder := NewGenkitEmbedder(genkit)

	// Create reranker if enabled
	var reranker Reranker
	if conf.RerankEnabled && embedder != nil {
		if conf.UseBatchRerank {
			reranker = NewBatchGenkitReranker(genkit, conf.RerankModel)
		} else {
			reranker = NewGenkitReranker(genkit, conf.RerankModel)
		}
	} else {
		reranker = NewNoOpReranker()
	}

	return &service{
		store:    store,
		embedder: embedder,
		reranker: reranker,
		config:   conf,
	}, nil
}

// NewServiceWithStore creates a new knowledge service with a custom knowledge store
func NewServiceWithStore(
	ctx context.Context,
	conf *config.KnowledgeConfig,
	logger *slog.Logger,
	genkit *genkit.Genkit,
	store Store,
) (Service, error) {
	// Create embedder for RAG functionality
	embedder := NewGenkitEmbedder(genkit)

	// Create reranker if enabled
	var reranker Reranker
	if conf.RerankEnabled && embedder != nil {
		if conf.UseBatchRerank {
			reranker = NewBatchGenkitReranker(genkit, conf.RerankModel)
		} else {
			reranker = NewGenkitReranker(genkit, conf.RerankModel)
		}
	} else {
		reranker = NewNoOpReranker()
	}

	return &service{
		store:    store,
		embedder: embedder,
		reranker: reranker,
		config:   conf,
	}, nil
}

func (s *service) Close() error {
	if s.store != nil {
		return s.store.Close()
	}
	return nil
}

// IndexKnowledge indexes knowledge documents for an agent
func (s *service) IndexKnowledge(ctx context.Context, agentName string, knowledge []map[string]any) error {
	if s.embedder == nil {
		// Return error instead of silently failing - this indicates a configuration issue
		return errors.New("embedder is not available - check OpenAI API key configuration. Knowledge indexing requires a valid OpenAI API key")
	}

	// First, delete existing knowledge for this agent
	if err := s.DeleteAgentKnowledge(ctx, agentName); err != nil {
		return errors.Wrapf(err, "failed to delete existing knowledge")
	}

	if len(knowledge) == 0 {
		return nil
	}

	// Process knowledge into text chunks
	chunks := s.processKnowledge(knowledge)
	if len(chunks) == 0 {
		return nil
	}

	// Extract text content for embedding
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	// Generate embeddings
	embeddings, err := s.embedder.Embed(ctx, texts...)
	if err != nil {
		return errors.Wrapf(err, "failed to generate embeddings")
	}

	if len(embeddings) != len(chunks) {
		return errors.Errorf("embedding count mismatch: got %d, expected %d", len(embeddings), len(chunks))
	}

	// Create knowledge items for storage
	now := time.Now()
	items := make([]KnowledgeItem, len(chunks))
	for i, chunk := range chunks {
		items[i] = KnowledgeItem{
			AgentName: agentName,
			Content:   chunk.Content,
			Embedding: embeddings[i],
			Metadata:  chunk.Metadata,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	// Store all items
	if err := s.store.Store(ctx, items); err != nil {
		return errors.Wrapf(err, "failed to store knowledge")
	}

	return nil
}

// RetrieveRelevantKnowledge retrieves relevant knowledge chunks based on query
func (s *service) RetrieveRelevantKnowledge(ctx context.Context, agentName string, query string, limit int) ([]string, error) {
	if s.embedder == nil {
		// Gracefully handle when no embedder is available
		return nil, nil
	}

	// Generate embedding for the query
	embeddings, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate query embedding")
	}

	if len(embeddings) == 0 {
		return nil, errors.Errorf("no embedding generated for query")
	}

	queryEmbedding := embeddings[0]

	// Determine retrieval count based on rerank configuration
	retrievalLimit := limit
	if s.config.RerankEnabled && s.config.RetrievalFactor > 1 {
		retrievalLimit = limit * s.config.RetrievalFactor
	}

	// Search for relevant knowledge
	searchResults, err := s.store.Search(ctx, agentName, queryEmbedding, retrievalLimit)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to search knowledge")
	}

	// Extract content for reranking
	candidates := make([]string, len(searchResults))
	for i, result := range searchResults {
		candidates[i] = result.Content
	}

	// Apply reranking if enabled
	if s.config.RerankEnabled && s.reranker != nil && len(candidates) > limit {
		rerankResults, err := s.reranker.Rerank(ctx, query, candidates, limit)
		if err != nil {
			// If reranking fails, fall back to original results
			logger := slog.Default()
			logger.Warn("reranking failed, falling back to original results", slog.String("error", err.Error()))
			if len(candidates) > limit {
				candidates = candidates[:limit]
			}
			return candidates, nil
		}

		// Extract content from rerank results
		results := make([]string, len(rerankResults))
		for i, result := range rerankResults {
			results[i] = result.Content
		}
		return results, nil
	}

	// If reranking is not enabled or not needed, return original results
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates, nil
}

// DeleteAgentKnowledge removes all knowledge for an agent
func (s *service) DeleteAgentKnowledge(ctx context.Context, agentName string) error {
	return s.store.DeleteByAgent(ctx, agentName)
}

// processKnowledge converts knowledge maps into indexable text chunks
func (s *service) processKnowledge(knowledge []map[string]any) []KnowledgeChunk {
	var chunks []KnowledgeChunk

	for _, item := range knowledge {
		// Convert the knowledge item to a searchable text representation
		content := s.extractTextFromKnowledge(item)
		if content == "" {
			continue
		}

		chunks = append(chunks, KnowledgeChunk{
			Content:  content,
			Metadata: item,
		})
	}

	return chunks
}

// extractTextFromKnowledge extracts searchable text from a knowledge map
func (s *service) extractTextFromKnowledge(item map[string]any) string {
	var textParts []string

	// Common text fields to extract (in priority order)
	textFields := []string{"content", "description", "title", "summary", "text", "name"}

	// First, look for standard text fields
	var foundStandardFields []string
	for _, field := range textFields {
		if value, exists := item[field]; exists {
			if str, ok := value.(string); ok && str != "" {
				foundStandardFields = append(foundStandardFields, str)
			}
		}
	}

	// If we found standard text fields, use them
	if len(foundStandardFields) > 0 {
		textParts = foundStandardFields
	} else {
		// If no standard text fields found, try to extract from all string values
		// Sort keys for deterministic ordering
		var keys []string
		for k := range item {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := item[key]
			if str, ok := value.(string); ok && str != "" {
				textParts = append(textParts, fmt.Sprintf("%s: %s", key, str))
			}
		}
	}

	return strings.Join(textParts, " ")
}
