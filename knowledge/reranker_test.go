package knowledge_test

import (
	"context"
	"testing"

	"github.com/habiliai/agentruntime/knowledge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockReranker is a mock implementation for testing
type MockReranker struct {
	rerankFunc func(ctx context.Context, query string, candidates []*knowledge.KnowledgeSearchResult, topK int) ([]*knowledge.KnowledgeSearchResult, error)
}

func (m *MockReranker) Rerank(ctx context.Context, query string, candidates []*knowledge.KnowledgeSearchResult, topK int) ([]*knowledge.KnowledgeSearchResult, error) {
	if m.rerankFunc != nil {
		return m.rerankFunc(ctx, query, candidates, topK)
	}
	return nil, nil
}

// Helper function to create KnowledgeSearchResult from string content
func createKnowledgeSearchResult(content string) *knowledge.KnowledgeSearchResult {
	return &knowledge.KnowledgeSearchResult{
		Document: &knowledge.Document{
			Content: knowledge.Content{
				MIMEType: "text/plain",
				Text:     content,
			},
		},
		Score: 1.0,
	}
}

// Helper function to get text content from KnowledgeSearchResult
func getTextContent(result *knowledge.KnowledgeSearchResult) string {
	if result.Document != nil && result.Document.Content.Type() == knowledge.ContentTypeText {
		return result.Document.Content.Text
	}
	return ""
}

func TestNoOpReranker(t *testing.T) {
	ctx := context.Background()
	reranker := knowledge.NewNoOpReranker()

	candidates := []*knowledge.KnowledgeSearchResult{
		createKnowledgeSearchResult("First candidate"),
		createKnowledgeSearchResult("Second candidate"),
		createKnowledgeSearchResult("Third candidate"),
		createKnowledgeSearchResult("Fourth candidate"),
		createKnowledgeSearchResult("Fifth candidate"),
	}

	t.Run("returns requested number of results", func(t *testing.T) {
		results, err := reranker.Rerank(ctx, "test query", candidates, 3)
		require.NoError(t, err)
		assert.Len(t, results, 3)

		// Check that results maintain original order
		assert.Equal(t, "First candidate", getTextContent(results[0]))
		assert.Equal(t, "Second candidate", getTextContent(results[1]))
		assert.Equal(t, "Third candidate", getTextContent(results[2]))

		// NoOpReranker doesn't modify scores
		for _, result := range results {
			assert.Equal(t, float32(1.0), result.Score)
		}
	})

	t.Run("handles topK larger than candidates", func(t *testing.T) {
		results, err := reranker.Rerank(ctx, "test query", candidates, 10)
		require.NoError(t, err)
		assert.Len(t, results, len(candidates))
	})

	t.Run("handles empty candidates", func(t *testing.T) {
		results, err := reranker.Rerank(ctx, "test query", []*knowledge.KnowledgeSearchResult{}, 5)
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestRerankResult(t *testing.T) {
	results := []*knowledge.KnowledgeSearchResult{
		{
			Document: &knowledge.Document{
				Content: knowledge.Content{
					MIMEType: "text/plain",
					Text:     "Low relevance",
				},
			},
			Score: 0.2,
		},
		{
			Document: &knowledge.Document{
				Content: knowledge.Content{
					MIMEType: "text/plain",
					Text:     "High relevance",
				},
			},
			Score: 0.9,
		},
		{
			Document: &knowledge.Document{
				Content: knowledge.Content{
					MIMEType: "text/plain",
					Text:     "Medium relevance",
				},
			},
			Score: 0.5,
		},
	}

	// Test manual sorting (like in the actual implementation)
	sorted := make([]*knowledge.KnowledgeSearchResult, len(results))
	copy(sorted, results)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Score < sorted[j].Score {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	assert.Equal(t, "High relevance", getTextContent(sorted[0]))
	assert.Equal(t, float32(0.9), sorted[0].Score)
	assert.Equal(t, "Medium relevance", getTextContent(sorted[1]))
	assert.Equal(t, float32(0.5), sorted[1].Score)
	assert.Equal(t, "Low relevance", getTextContent(sorted[2]))
	assert.Equal(t, float32(0.2), sorted[2].Score)
}

func TestRerankWithMock(t *testing.T) {
	t.Run("reranker filters and reorders results", func(t *testing.T) {
		ctx := context.Background()

		// Mock reranker that assigns scores based on content
		mockReranker := &MockReranker{
			rerankFunc: func(ctx context.Context, query string, candidates []*knowledge.KnowledgeSearchResult, topK int) ([]*knowledge.KnowledgeSearchResult, error) {
				results := make([]*knowledge.KnowledgeSearchResult, 0, len(candidates))

				for _, candidate := range candidates {
					content := getTextContent(candidate)
					var score float32
					switch content {
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

					// Create new result with updated score
					results = append(results, &knowledge.KnowledgeSearchResult{
						Document: candidate.Document,
						Score:    score,
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

		candidates := []*knowledge.KnowledgeSearchResult{
			createKnowledgeSearchResult("Not relevant"),
			createKnowledgeSearchResult("Highly relevant"),
			createKnowledgeSearchResult("Barely relevant"),
			createKnowledgeSearchResult("Somewhat relevant"),
		}

		results, err := mockReranker.Rerank(ctx, "test query", candidates, 2)
		require.NoError(t, err)
		require.Len(t, results, 2)

		// Should return top 2 by score
		assert.Equal(t, "Highly relevant", getTextContent(results[0]))
		assert.Equal(t, float32(0.95), results[0].Score)
		assert.Equal(t, "Somewhat relevant", getTextContent(results[1]))
		assert.Equal(t, float32(0.6), results[1].Score)
	})
}
