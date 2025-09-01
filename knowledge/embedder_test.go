package knowledge

import (
	_ "embed"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

var (
	//go:embed testdata/gen_image_1.png
	genImage1 []byte

	//go:embed testdata/gen_image_2.png
	genImage2 []byte
)

func TestEmbeddingTaskType_String(t *testing.T) {
	tests := []struct {
		name     string
		taskType EmbeddingTaskType
		expected string
	}{
		{
			name:     "Document task type",
			taskType: EmbeddingTaskTypeDocument,
			expected: "search_document",
		},
		{
			name:     "Query task type",
			taskType: EmbeddingTaskTypeQuery,
			expected: "search_query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.taskType.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEmbedder_GetEmbedSize(t *testing.T) {
	embedder := NewEmbedder("test-key")
	expected := 768
	result := embedder.GetEmbedSize()

	if result != expected {
		t.Errorf("expected %d, got %d", expected, result)
	}
}

func TestEmbedder_EmbedTexts(t *testing.T) {
	// Load .env file if it exists (try both current directory and parent directory)
	_ = godotenv.Load("../.env")

	apiKey := os.Getenv("NOMIC_API_KEY")
	if apiKey == "" {
		t.Skip("NOMIC_API_KEY environment variable not set, skipping live API test")
	}

	embedder := NewEmbedder(apiKey)

	t.Run("successful embedding", func(t *testing.T) {
		embeddings, err := embedder.EmbedTexts(t.Context(), EmbeddingTaskTypeDocument, "hello", "world")

		require.NoError(t, err)
		require.Len(t, embeddings, 2)

		// Verify each embedding has the correct dimension
		for i, embedding := range embeddings {
			require.Len(t, embedding, 768, "embedding %d has wrong dimension", i)
		}

		// Verify embeddings are different (they should be for different texts)
		require.GreaterOrEqual(t, len(embeddings), 2)
		same := true
		for i := range embeddings[0] {
			if embeddings[0][i] != embeddings[1][i] {
				same = false
				break
			}
		}
		require.False(t, same, "embeddings for 'hello' and 'world' are identical, expected different embeddings")
	})

	t.Run("single text embedding", func(t *testing.T) {
		embeddings, err := embedder.EmbedTexts(t.Context(), EmbeddingTaskTypeQuery, "test query")

		require.NoError(t, err)
		require.Len(t, embeddings, 1)
		require.Len(t, embeddings[0], 768)
	})
}

func TestEmbedder_EmbedImageUrls(t *testing.T) {
	// Load .env file if it exists (try both current directory and parent directory)
	_ = godotenv.Load("../.env")

	apiKey := os.Getenv("NOMIC_API_KEY")
	if apiKey == "" {
		t.Skip("NOMIC_API_KEY environment variable not set, skipping live API test")
	}

	embedder := NewEmbedder(apiKey)

	t.Run("successful embedding with public image", func(t *testing.T) {
		// Using Nomic's example image URL from their documentation
		imageURL := "https://static.nomic.ai/secret-model.png"

		embeddings, err := embedder.EmbedImageUrls(t.Context(), imageURL)

		require.NoError(t, err)
		require.Len(t, embeddings, 1)
		require.Len(t, embeddings[0], 768)

		// Verify embedding is not all zeros
		allZeros := true
		for _, val := range embeddings[0] {
			if val != 0 {
				allZeros = false
				break
			}
		}
		require.False(t, allZeros, "embedding is all zeros, expected non-zero values")
	})

	t.Run("multiple image URLs", func(t *testing.T) {
		// Using Nomic's example image URLs from their documentation
		imageURLs := []string{
			"https://static.nomic.ai/secret-model.png",
			"https://static.nomic.ai/secret-model-2.png",
		}

		embeddings, err := embedder.EmbedImageUrls(t.Context(), imageURLs...)

		require.NoError(t, err)
		require.Len(t, embeddings, 2)

		// Verify each embedding has the correct dimension
		for i, embedding := range embeddings {
			require.Len(t, embedding, 768, "embedding %d has wrong dimension", i)
		}
	})
}

func TestEmbedder_EmbedImageFiles(t *testing.T) {
	// Load .env file if it exists (try both current directory and parent directory)
	_ = godotenv.Load("../.env")

	apiKey := os.Getenv("NOMIC_API_KEY")
	if apiKey == "" {
		t.Skip("NOMIC_API_KEY environment variable not set, skipping live API test")
	}

	embedder := NewEmbedder(apiKey)

	t.Run("successful embedding with small PNG", func(t *testing.T) {
		// Create a minimal 1x1 PNG image in memory
		// PNG signature + IHDR chunk + IDAT chunk + IEND chunk

		embeddings, err := embedder.EmbedImageFiles(t.Context(), "image/png", genImage1)

		require.NoError(t, err)
		require.Len(t, embeddings, 1)
		require.Len(t, embeddings[0], 768)

		// Verify embedding is not all zeros
		allZeros := true
		for _, val := range embeddings[0] {
			if val != 0 {
				allZeros = false
				break
			}
		}
		require.False(t, allZeros, "embedding is all zeros, expected non-zero values")
	})

	t.Run("multiple image files", func(t *testing.T) {
		embeddings, err := embedder.EmbedImageFiles(t.Context(), "image/png", genImage1, genImage2)

		require.NoError(t, err)
		require.Len(t, embeddings, 2)

		// Verify each embedding has the correct dimension
		for i, embedding := range embeddings {
			require.Len(t, embedding, 768, "embedding %d has wrong dimension", i)
		}
	})

	t.Run("single page image file", func(t *testing.T) {
		pageImage, err := os.ReadFile("testdata/page_1.jpg")
		require.NoError(t, err)

		embeddings, err := embedder.EmbedImageFiles(t.Context(), "image/jpeg", pageImage)

		require.NoError(t, err)
		require.Len(t, embeddings, 1)
		require.Len(t, embeddings[0], 768)
	})
}
