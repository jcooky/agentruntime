package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// mockEmbedder for testing
type mockEmbedder struct{}

func (m *mockEmbedder) Embed(ctx context.Context, texts ...string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		// Create simple mock embeddings (in real usage, these would be from OpenAI)
		embedding := make([]float32, 1536)
		for j := range embedding {
			embedding[j] = float32(i+j) * 0.001 // Simple pattern for testing
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}

func TestKnowledgeProcessing(t *testing.T) {
	service := &SqliteService{embedder: &mockEmbedder{}}

	knowledge := []map[string]any{
		{
			"cityName": "Seoul",
			"aliases":  "Seoul, SEOUL, KOR, Korea",
			"info":     "Capital city of South Korea, known for technology and K-pop culture",
			"weather":  "Four distinct seasons with hot summers and cold winters",
		},
		{
			"cityName": "Tokyo",
			"aliases":  "Tokyo, TYO, Japan",
			"info":     "Capital city of Japan, largest metropolitan area in the world",
			"weather":  "Humid subtropical climate with hot, humid summers",
		},
	}

	chunks := service.processKnowledge(knowledge)
	require.Len(t, chunks, 2)

	// Check that content is extracted properly
	require.Contains(t, chunks[0].Content, "Seoul")
	require.Contains(t, chunks[0].Content, "South Korea")
	require.Contains(t, chunks[1].Content, "Tokyo")
	require.Contains(t, chunks[1].Content, "Japan")

	// Check that metadata is preserved
	require.Equal(t, knowledge[0], chunks[0].Metadata)
	require.Equal(t, knowledge[1], chunks[1].Metadata)
}

func TestTextExtraction(t *testing.T) {
	service := &SqliteService{}

	testCases := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{
			name: "standard text fields",
			input: map[string]any{
				"title":       "Test Title",
				"description": "Test Description",
				"content":     "Test Content",
			},
			expected: "Test Content Test Description Test Title",
		},
		{
			name: "custom fields",
			input: map[string]any{
				"cityName": "Seoul",
				"country":  "South Korea",
				"info":     "Technology hub",
			},
			expected: "cityName: Seoul country: South Korea info: Technology hub",
		},
		{
			name: "mixed types",
			input: map[string]any{
				"name":        "Test",
				"count":       42,
				"active":      true,
				"description": "Valid text",
			},
			expected: "Valid text Test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.extractTextFromKnowledge(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
