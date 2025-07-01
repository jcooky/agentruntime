package knowledge_test

import (
	"context"
	"testing"

	"github.com/habiliai/agentruntime/knowledge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryStore_StoreAndRetrieve(t *testing.T) {
	ctx := context.Background()
	store := knowledge.NewInMemoryStore()
	defer store.Close()

	// Create test knowledge
	testKnowledge := &knowledge.Knowledge{
		ID: "test-knowledge-1",
		Source: knowledge.Source{
			Title: "Test Source",
			Type:  knowledge.SourceTypeMap,
		},
		Metadata: map[string]any{
			"test": "metadata",
		},
		Documents: []*knowledge.Document{
			{
				ID: "doc-1",
				Content: knowledge.Content{
					Type: knowledge.ContentTypeText,
					Text: "This is a test document about Python programming",
				},
				EmbeddingText: "This is a test document about Python programming",
				Embeddings:    generateTestEmbedding(128, 1), // 128-dim embedding with seed 1
				Metadata: map[string]any{
					"language": "python",
				},
			},
			{
				ID: "doc-2",
				Content: knowledge.Content{
					Type: knowledge.ContentTypeText,
					Text: "JavaScript is a dynamic programming language",
				},
				EmbeddingText: "JavaScript is a dynamic programming language",
				Embeddings:    generateTestEmbedding(128, 2), // 128-dim embedding with seed 2
				Metadata: map[string]any{
					"language": "javascript",
				},
			},
		},
	}

	// Test Store
	err := store.Store(ctx, testKnowledge)
	require.NoError(t, err)

	// Test GetKnowledgeById
	retrieved, err := store.GetKnowledgeById(ctx, "test-knowledge-1")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, testKnowledge.ID, retrieved.ID)
	assert.Equal(t, testKnowledge.Source.Title, retrieved.Source.Title)
	assert.Equal(t, len(testKnowledge.Documents), len(retrieved.Documents))
	assert.Equal(t, "python", retrieved.Documents[0].Metadata["language"])
	assert.Equal(t, "javascript", retrieved.Documents[1].Metadata["language"])

	// Test GetKnowledgeById with non-existent ID
	notFound, err := store.GetKnowledgeById(ctx, "non-existent")
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestInMemoryStore_Search(t *testing.T) {
	ctx := context.Background()
	store := knowledge.NewInMemoryStore()
	defer store.Close()

	// Store multiple documents with different embeddings
	documents := []*knowledge.Document{
		{
			ID: "doc-1",
			Content: knowledge.Content{
				Type: knowledge.ContentTypeText,
				Text: "Python programming basics",
			},
			EmbeddingText: "Python programming basics",
			Embeddings:    generateTestEmbedding(128, 1),
		},
		{
			ID: "doc-2",
			Content: knowledge.Content{
				Type: knowledge.ContentTypeText,
				Text: "Advanced Python techniques",
			},
			EmbeddingText: "Advanced Python techniques",
			Embeddings:    generateTestEmbedding(128, 2),
		},
		{
			ID: "doc-3",
			Content: knowledge.Content{
				Type: knowledge.ContentTypeText,
				Text: "JavaScript fundamentals",
			},
			EmbeddingText: "JavaScript fundamentals",
			Embeddings:    generateTestEmbedding(128, 3),
		},
	}

	for i, doc := range documents {
		knowledge := &knowledge.Knowledge{
			ID:        string(rune('a' + i)),
			Documents: []*knowledge.Document{doc},
		}
		err := store.Store(ctx, knowledge)
		require.NoError(t, err)
	}

	// Search with a query embedding similar to doc-1
	queryEmbedding := generateTestEmbedding(128, 1) // Same as doc-1
	results, err := store.Search(ctx, queryEmbedding, 2)
	require.NoError(t, err)
	require.Len(t, results, 2)

	// First result should be doc-1 (exact match)
	assert.Equal(t, "doc-1", results[0].Document.ID)
	assert.Greater(t, results[0].Score, float32(0.99)) // Should be very close to 1.0

	// Test with empty query embedding
	emptyResults, err := store.Search(ctx, []float32{}, 10)
	require.NoError(t, err)
	assert.Empty(t, emptyResults)
}

func TestInMemoryStore_Delete(t *testing.T) {
	ctx := context.Background()
	store := knowledge.NewInMemoryStore()
	defer store.Close()

	// Store knowledge
	knowledge := &knowledge.Knowledge{
		ID: "test-knowledge",
		Documents: []*knowledge.Document{
			{
				ID: "doc-1",
				Content: knowledge.Content{
					Type: knowledge.ContentTypeText,
					Text: "Test document",
				},
				EmbeddingText: "Test document",
				Embeddings:    generateTestEmbedding(128, 1),
			},
		},
	}
	err := store.Store(ctx, knowledge)
	require.NoError(t, err)

	// Verify it exists
	retrieved, err := store.GetKnowledgeById(ctx, "test-knowledge")
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	// Delete knowledge
	err = store.DeleteKnowledgeById(ctx, "test-knowledge")
	require.NoError(t, err)

	// Verify it's deleted
	deleted, err := store.GetKnowledgeById(ctx, "test-knowledge")
	require.NoError(t, err)
	assert.Nil(t, deleted)

	// Delete non-existent knowledge should not error
	err = store.DeleteKnowledgeById(ctx, "non-existent")
	require.NoError(t, err)
}

func TestInMemoryStore_Concurrency(t *testing.T) {
	ctx := context.Background()
	store := knowledge.NewInMemoryStore()
	defer store.Close()

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			knowledge := &knowledge.Knowledge{
				ID: string(rune('a' + id)),
				Documents: []*knowledge.Document{
					{
						ID: string(rune('a'+id)) + "-doc",
						Content: knowledge.Content{
							Type: knowledge.ContentTypeText,
							Text: "Concurrent test",
						},
						EmbeddingText: "Concurrent test",
						Embeddings:    generateTestEmbedding(128, id),
					},
				},
			}
			err := store.Store(ctx, knowledge)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all were stored
	for i := 0; i < 10; i++ {
		knowledge, err := store.GetKnowledgeById(ctx, string(rune('a'+i)))
		require.NoError(t, err)
		require.NotNil(t, knowledge)
	}
}

func TestInMemoryStore_DeepCopy(t *testing.T) {
	ctx := context.Background()
	store := knowledge.NewInMemoryStore()
	defer store.Close()

	// Create original knowledge
	original := &knowledge.Knowledge{
		ID: "test",
		Metadata: map[string]any{
			"key": "value",
		},
		Documents: []*knowledge.Document{
			{
				ID: "doc-1",
				Metadata: map[string]any{
					"doc-key": "doc-value",
				},
			},
		},
	}

	// Store it
	err := store.Store(ctx, original)
	require.NoError(t, err)

	// Modify original after storing
	original.Metadata["key"] = "modified"
	original.Documents[0].Metadata["doc-key"] = "modified"

	// Retrieve and check that stored version is unchanged
	retrieved, err := store.GetKnowledgeById(ctx, "test")
	require.NoError(t, err)
	assert.Equal(t, "value", retrieved.Metadata["key"])
	assert.Equal(t, "doc-value", retrieved.Documents[0].Metadata["doc-key"])
}

// Helper function to generate deterministic test embeddings
func generateTestEmbedding(dim int, seed int) []float32 {
	embedding := make([]float32, dim)
	for i := 0; i < dim; i++ {
		// Simple deterministic function
		embedding[i] = float32((seed+i)%10) / 10.0
	}
	return embedding
}
