package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockReranker is a mock implementation for testing
type MockReranker struct {
	rerankFunc func(ctx context.Context, query string, candidates []string, topK int) ([]RerankResult, error)
}

func (m *MockReranker) Rerank(ctx context.Context, query string, candidates []string, topK int) ([]RerankResult, error) {
	if m.rerankFunc != nil {
		return m.rerankFunc(ctx, query, candidates, topK)
	}
	return nil, nil
}

func TestNoOpReranker(t *testing.T) {
	ctx := context.Background()
	reranker := NewNoOpReranker()

	candidates := []string{
		"First candidate",
		"Second candidate",
		"Third candidate",
		"Fourth candidate",
		"Fifth candidate",
	}

	t.Run("returns requested number of results", func(t *testing.T) {
		results, err := reranker.Rerank(ctx, "test query", candidates, 3)
		require.NoError(t, err)
		assert.Len(t, results, 3)

		// Check that results maintain original order
		assert.Equal(t, "First candidate", results[0].Content)
		assert.Equal(t, "Second candidate", results[1].Content)
		assert.Equal(t, "Third candidate", results[2].Content)

		// All should have the same score
		for _, result := range results {
			assert.Equal(t, 1.0, result.Score)
		}
	})

	t.Run("handles topK larger than candidates", func(t *testing.T) {
		results, err := reranker.Rerank(ctx, "test query", candidates, 10)
		require.NoError(t, err)
		assert.Len(t, results, len(candidates))
	})

	t.Run("handles empty candidates", func(t *testing.T) {
		results, err := reranker.Rerank(ctx, "test query", []string{}, 5)
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestRerankResult(t *testing.T) {
	results := []RerankResult{
		{Content: "Low relevance", Score: 0.2},
		{Content: "High relevance", Score: 0.9},
		{Content: "Medium relevance", Score: 0.5},
	}

	// Test manual sorting (like in the actual implementation)
	sorted := make([]RerankResult, len(results))
	copy(sorted, results)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Score < sorted[j].Score {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	assert.Equal(t, "High relevance", sorted[0].Content)
	assert.Equal(t, 0.9, sorted[0].Score)
	assert.Equal(t, "Medium relevance", sorted[1].Content)
	assert.Equal(t, 0.5, sorted[1].Score)
	assert.Equal(t, "Low relevance", sorted[2].Content)
	assert.Equal(t, 0.2, sorted[2].Score)
}

func TestRerankWithMock(t *testing.T) {
	t.Run("reranker filters and reorders results", func(t *testing.T) {
		ctx := context.Background()

		// Mock reranker that assigns scores based on content
		mockReranker := &MockReranker{
			rerankFunc: func(ctx context.Context, query string, candidates []string, topK int) ([]RerankResult, error) {
				results := make([]RerankResult, 0, len(candidates))

				for _, candidate := range candidates {
					var score float64
					switch candidate {
					case "Highly relevant":
						score = 0.95
					case "Somewhat relevant":
						score = 0.6
					case "Barely relevant":
						score = 0.3
					case "Not relevant":
						score = 0.1
					default:
						score = 0.5
					}

					results = append(results, RerankResult{
						Content: candidate,
						Score:   score,
					})
				}

				// Sort by score
				for i := 0; i < len(results)-1; i++ {
					for j := i + 1; j < len(results); j++ {
						if results[i].Score < results[j].Score {
							results[i], results[j] = results[j], results[i]
						}
					}
				}

				if topK < len(results) {
					results = results[:topK]
				}

				return results, nil
			},
		}

		candidates := []string{
			"Not relevant",
			"Highly relevant",
			"Barely relevant",
			"Somewhat relevant",
		}

		results, err := mockReranker.Rerank(ctx, "test query", candidates, 2)
		require.NoError(t, err)
		require.Len(t, results, 2)

		// Should return top 2 by score
		assert.Equal(t, "Highly relevant", results[0].Content)
		assert.Equal(t, 0.95, results[0].Score)
		assert.Equal(t, "Somewhat relevant", results[1].Content)
		assert.Equal(t, 0.6, results[1].Score)
	})
}
