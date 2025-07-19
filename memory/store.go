package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/pkg/errors"
	"gonum.org/v1/gonum/mat"
)

type (
	// Store interface for memory storage
	Store interface {
		Set(ctx context.Context, memory *Memory) error
		Get(ctx context.Context, key string) (*Memory, error)
		Search(ctx context.Context, query string, queryEmbedding []float32, limit uint) ([]ScoredMemory, error)
		List(ctx context.Context) ([]*Memory, error)
		Delete(ctx context.Context, key string) error
	}

	// InMemoryStore is a simple in-memory implementation
	InMemoryStore struct {
		mu       sync.RWMutex
		memories map[string]*Memory
	}
)

// Memory tool functions
var (
	_ Store = (*InMemoryStore)(nil)
)

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		memories: make(map[string]*Memory),
	}
}

func (s *InMemoryStore) Set(ctx context.Context, memory *Memory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.memories[memory.Key]; exists {
		return errors.Errorf("memory with key '%s' already exists", memory.Key)
	}

	s.memories[memory.Key] = memory
	return nil
}

func (s *InMemoryStore) Get(ctx context.Context, key string) (*Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	memory, exists := s.memories[key]
	if !exists {
		return nil, errors.Errorf("memory with key '%s' not found", key)
	}
	return memory, nil
}

func (s *InMemoryStore) Search(ctx context.Context, query string, queryEmbedding []float32, limit uint) ([]ScoredMemory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(queryEmbedding) == 0 {
		return nil, errors.New("query embedding is empty")
	}

	// Collect all memories
	var validMemories []*Memory
	for _, memory := range s.memories {
		// Only include memories with matching embedding dimensions for matrix calculation
		if len(queryEmbedding) > 0 && len(memory.Embedding) == len(queryEmbedding) {
			validMemories = append(validMemories, memory)
		}
	}

	if len(validMemories) == 0 {
		return nil, errors.New("no memories found")
	}

	// Create scored results for all memories
	scoredResults := make([]ScoredMemory, 0, len(validMemories))

	// Calculate scores for valid memories using matrix multiplication
	numMemories := len(validMemories)
	embeddingDim := len(queryEmbedding)

	// Convert queryEmbedding to float64 vector
	queryVec := make([]float64, embeddingDim)
	for i, v := range queryEmbedding {
		queryVec[i] = float64(v)
	}

	// Create memory embeddings matrix (N x d)
	memoryData := make([]float64, numMemories*embeddingDim)
	for i, memory := range validMemories {
		for j, v := range memory.Embedding {
			memoryData[i*embeddingDim+j] = float64(v)
		}
	}

	// Create gonum matrices
	queryVector := mat.NewVecDense(embeddingDim, queryVec)
	memoryMatrix := mat.NewDense(numMemories, embeddingDim, memoryData)

	// Perform matrix multiplication: memoryMatrix * queryVector = similarity scores
	var resultVec mat.VecDense
	resultVec.MulVec(memoryMatrix, queryVector)

	// Create scored results with OpenAI embedding optimized transformation
	// OpenAI embeddings are normalized, so inner product is always in [-1, 1]
	// Transform to [0, 1] range: (score + 1) * 0.5
	for i, memory := range validMemories {
		score := (resultVec.AtVec(i) + 1.0) * 0.5 // [-1,1] → [0,1]

		scoredResults = append(scoredResults, ScoredMemory{
			Memory: memory,
			Score:  score,
		})
	}

	// Sort by similarity score (descending)
	sort.Slice(scoredResults, func(i, j int) bool {
		return scoredResults[i].Score > scoredResults[j].Score
	})

	// Apply limit
	if limit > 0 && uint(len(scoredResults)) > limit {
		scoredResults = scoredResults[:limit]
	}

	return scoredResults, nil
}

func (s *InMemoryStore) List(ctx context.Context) ([]*Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*Memory, 0, len(s.memories))
	for _, memory := range s.memories {
		results = append(results, memory)
	}

	return results, nil
}

func (s *InMemoryStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.memories, key)
	return nil
}
