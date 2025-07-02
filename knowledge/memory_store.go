package knowledge

import (
	"context"
	"math"
	"slices"
	"sort"
	"sync"
)

type (
	InMemoryStore struct {
		mu         sync.RWMutex
		knowledges map[string]*Knowledge // key: knowledge ID
	}
)

// NewInMemoryStore creates a new in-memory knowledge store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		knowledges: make(map[string]*Knowledge),
	}
}

// Store implements Store.Store
func (i *InMemoryStore) Store(ctx context.Context, knowledge *Knowledge) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Deep copy the knowledge to avoid external modifications
	storedKnowledge := &Knowledge{
		ID:        knowledge.ID,
		Source:    knowledge.Source,
		Metadata:  copyMap(knowledge.Metadata),
		Documents: make([]*Document, len(knowledge.Documents)),
	}

	// Store knowledge
	i.knowledges[knowledge.ID] = storedKnowledge

	// Store documents
	for idx, doc := range knowledge.Documents {
		// Deep copy document
		storedDoc := &Document{
			ID:            doc.ID,
			Content:       doc.Content, // Contents are immutable, so shallow copy is ok
			Embeddings:    copyFloat32Slice(doc.Embeddings),
			EmbeddingText: doc.EmbeddingText,
			Metadata:      copyMap(doc.Metadata),
		}

		// Add knowledge ID to document metadata for reference
		if storedDoc.Metadata == nil {
			storedDoc.Metadata = make(map[string]any)
		}
		storedDoc.Metadata["knowledge_id"] = knowledge.ID

		storedKnowledge.Documents[idx] = storedDoc
	}

	return nil
}

// Search implements Store.Search
func (i *InMemoryStore) Search(ctx context.Context, queryEmbedding []float32, limit int, allowedKnowledgeIds []string) ([]KnowledgeSearchResult, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if len(queryEmbedding) == 0 {
		return []KnowledgeSearchResult{}, nil
	}

	// Calculate cosine similarity for all documents with embeddings
	type scoredDoc struct {
		doc   *Document
		score float32
	}

	var scoredDocs []scoredDoc
	for _, kl := range i.knowledges {
		if len(allowedKnowledgeIds) > 0 && !slices.Contains(allowedKnowledgeIds, kl.ID) {
			continue
		}
		for _, doc := range kl.Documents {
			if len(doc.Embeddings) == 0 {
				continue
			}

			// Calculate cosine similarity
			similarity := cosineSimilarity(queryEmbedding, doc.Embeddings)
			scoredDocs = append(scoredDocs, scoredDoc{
				doc:   doc,
				score: similarity,
			})
		}
	}

	// Sort by score descending
	sort.Slice(scoredDocs, func(i, j int) bool {
		return scoredDocs[i].score > scoredDocs[j].score
	})

	// Limit results
	if len(scoredDocs) > limit {
		scoredDocs = scoredDocs[:limit]
	}

	// Convert to search results
	results := make([]KnowledgeSearchResult, len(scoredDocs))
	for i, sd := range scoredDocs {
		// Deep copy document for result
		resultDoc := &Document{
			ID:            sd.doc.ID,
			Content:       sd.doc.Content,
			Embeddings:    nil, // Don't include embeddings in search results
			EmbeddingText: sd.doc.EmbeddingText,
			Metadata:      copyMap(sd.doc.Metadata),
		}

		results[i] = KnowledgeSearchResult{
			Document: resultDoc,
			Score:    sd.score,
		}
	}

	return results, nil
}

// GetKnowledgeById implements Store.GetKnowledgeById
func (i *InMemoryStore) GetKnowledgeById(ctx context.Context, knowledgeId string) (*Knowledge, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	knowledge, exists := i.knowledges[knowledgeId]
	if !exists {
		return nil, nil // Return nil, nil for not found (consistent with SQLite store)
	}

	// Deep copy knowledge to avoid external modifications
	result := &Knowledge{
		ID:        knowledge.ID,
		Source:    knowledge.Source,
		Metadata:  copyMap(knowledge.Metadata),
		Documents: make([]*Document, len(knowledge.Documents)),
	}

	// Deep copy documents
	for idx, doc := range knowledge.Documents {
		result.Documents[idx] = &Document{
			ID:            doc.ID,
			Content:       doc.Content,
			Embeddings:    nil, // Don't include embeddings in get results
			EmbeddingText: doc.EmbeddingText,
			Metadata:      copyMap(doc.Metadata),
		}
	}

	return result, nil
}

// DeleteKnowledgeById implements Store.DeleteKnowledgeById
func (i *InMemoryStore) DeleteKnowledgeById(ctx context.Context, knowledgeId string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, exists := i.knowledges[knowledgeId]; !exists {
		return nil // Not an error if knowledge doesn't exist
	}

	// Delete knowledge
	delete(i.knowledges, knowledgeId)

	return nil
}

// Close implements Store.Close
func (i *InMemoryStore) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Clear all data
	i.knowledges = make(map[string]*Knowledge)

	return nil
}

// Helper function to calculate cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}

// Helper function to deep copy a map
func copyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}

	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// Helper function to copy a float32 slice
func copyFloat32Slice(s []float32) []float32 {
	if s == nil {
		return nil
	}

	result := make([]float32, len(s))
	copy(result, s)
	return result
}

var (
	_ Store = (*InMemoryStore)(nil)
)
