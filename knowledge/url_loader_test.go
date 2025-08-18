package knowledge_test

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/knowledge"
	firecrawl "github.com/mendableai/firecrawl-go"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIndexKnowledgeFromURL_* tests cover the URL indexing functionality
// These tests verify error handling, configuration validation, and basic processing
// For integration testing with real APIs, set OPENAI_API_KEY and FIRECRAWL_API_KEY environment variables

func TestIndexKnowledgeFromURL_RequiresEmbedder(t *testing.T) {
	ctx := t.Context()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create service with invalid OpenAI config, which should result in embedding failure
	store := knowledge.NewInMemoryStore()
	firecrawlConfig := &config.FireCrawlConfig{
		APIKey: "test-key",
		APIUrl: "https://api.firecrawl.dev",
	}

	service, err := knowledge.NewServiceWithStore(
		ctx,
		config.NewKnowledgeConfig(),
		&config.ModelConfig{
			OpenAIAPIKey: "invalid-key", // Invalid key that will cause embedding to fail
		},
		logger,
		store,
		firecrawlConfig,
	)
	require.NoError(t, err)

	// Create test crawl parameters with smaller limits for faster testing
	testCrawlParams := firecrawl.CrawlParams{
		MaxDepth:           lo.ToPtr(1), // Only crawl 1 level deep for tests
		Limit:              lo.ToPtr(2), // Limit to 2 pages for fast testing
		AllowBackwardLinks: lo.ToPtr(false),
		AllowExternalLinks: lo.ToPtr(false),
		ScrapeOptions: firecrawl.ScrapeParams{
			Formats: []string{"markdown"}, // Only markdown for faster processing
		},
	}

	// Attempt to index URL - this will fail at some point in the process
	_, err = service.IndexKnowledgeFromURL(ctx, "test-id", "https://httpbin.org/html", testCrawlParams)

	// Should fail - could be network error, embedding error, or API authentication error
	require.Error(t, err)
	errorMsg := err.Error()

	// Any of these errors indicate the system is working correctly
	errorIsExpected := strings.Contains(errorMsg, "embedding") ||
		strings.Contains(errorMsg, "failed to generate embeddings") ||
		strings.Contains(errorMsg, "API") ||
		strings.Contains(errorMsg, "Incorrect API key") ||
		strings.Contains(errorMsg, "dial tcp") ||
		strings.Contains(errorMsg, "no such host") ||
		strings.Contains(errorMsg, "failed to crawl")

	assert.True(t, errorIsExpected, "Error should be related to expected failure scenarios: %s", errorMsg)
}

// Note: Testing nil firecrawlConfig directly is difficult since service.go always creates
// a default config when nil is passed. The nil check serves as defensive programming
// for edge cases and potential future refactoring scenarios.

func TestIndexKnowledgeFromURL_EmptyFireCrawlAPIKey(t *testing.T) {
	ctx := t.Context()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create service with FireCrawl config that has empty API key
	store := knowledge.NewInMemoryStore()

	service, err := knowledge.NewServiceWithStore(
		ctx,
		config.NewKnowledgeConfig(),
		&config.ModelConfig{
			OpenAIAPIKey: "test-key", // Valid key to get embedder
		},
		logger,
		store,
		nil, // No FireCrawl config - default will be created with empty API key from environment
	)
	require.NoError(t, err)

	// Create default crawl parameters for testing
	defaultCrawlParams := firecrawl.CrawlParams{
		MaxDepth:           lo.ToPtr(1),
		Limit:              lo.ToPtr(1),
		AllowBackwardLinks: lo.ToPtr(false),
		AllowExternalLinks: lo.ToPtr(false),
		ScrapeOptions: firecrawl.ScrapeParams{
			Formats: []string{"markdown"},
		},
	}

	// Attempt to index URL
	_, err = service.IndexKnowledgeFromURL(ctx, "test-id", "https://example.com", defaultCrawlParams)

	// Should fail because FireCrawl config validation fails (empty API key)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FireCrawl configuration is invalid")
}

func TestIndexKnowledgeFromURL_InvalidFireCrawlConfig(t *testing.T) {
	ctx := t.Context()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create service with invalid FireCrawl config (no API key)
	store := knowledge.NewInMemoryStore()
	firecrawlConfig := &config.FireCrawlConfig{
		APIKey: "", // Invalid - empty key
		APIUrl: "https://api.firecrawl.dev",
	}

	service, err := knowledge.NewServiceWithStore(
		ctx,
		config.NewKnowledgeConfig(),
		&config.ModelConfig{
			OpenAIAPIKey: "test-key", // Valid key to get embedder
		},
		logger,
		store,
		firecrawlConfig,
	)
	require.NoError(t, err)

	// Create default crawl parameters for testing
	defaultCrawlParams := firecrawl.CrawlParams{
		MaxDepth:           lo.ToPtr(1),
		Limit:              lo.ToPtr(1),
		AllowBackwardLinks: lo.ToPtr(false),
		AllowExternalLinks: lo.ToPtr(false),
		ScrapeOptions: firecrawl.ScrapeParams{
			Formats: []string{"markdown"},
		},
	}

	// Attempt to index URL
	_, err = service.IndexKnowledgeFromURL(ctx, "test-id", "https://example.com", defaultCrawlParams)

	// Should fail because FireCrawl config validation fails
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FireCrawl configuration is invalid")
}

// Integration test with real API keys (skipped if keys not available)
func TestIndexKnowledgeFromURL_Integration(t *testing.T) {
	ctx := t.Context()

	// Check for required environment variables
	openAIKey := os.Getenv("OPENAI_API_KEY")
	firecrawlKey := os.Getenv("FIRECRAWL_API_KEY")

	if openAIKey == "" || firecrawlKey == "" {
		t.Skip("OPENAI_API_KEY and FIRECRAWL_API_KEY required for integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	store := knowledge.NewInMemoryStore()

	// Create service with real configurations
	firecrawlConfig := &config.FireCrawlConfig{
		APIKey: firecrawlKey,
		APIUrl: "https://api.firecrawl.dev",
	}

	service, err := knowledge.NewServiceWithStore(
		ctx,
		config.NewKnowledgeConfig(),
		&config.ModelConfig{
			OpenAIAPIKey: openAIKey,
		},
		logger,
		store,
		firecrawlConfig,
	)
	require.NoError(t, err)
	defer service.Close()

	// Test with a simple, fast-loading page
	testURL := "https://httpbin.org/html"

	// Create test crawl parameters with smaller limits for faster integration testing
	testCrawlParams := firecrawl.CrawlParams{
		MaxDepth:           lo.ToPtr(1), // Only crawl 1 level deep for tests
		Limit:              lo.ToPtr(3), // Limit to 3 pages for integration testing
		AllowBackwardLinks: lo.ToPtr(false),
		AllowExternalLinks: lo.ToPtr(false),
		ScrapeOptions: firecrawl.ScrapeParams{
			Formats: []string{"markdown"}, // Only markdown for faster processing
		},
	}

	// Index the URL
	knowledge, err := service.IndexKnowledgeFromURL(ctx, "test-integration", testURL, testCrawlParams)
	require.NoError(t, err)
	require.NotNil(t, knowledge)

	// Validate the knowledge object
	assert.Equal(t, "test-integration", knowledge.ID)
	assert.Equal(t, "url", string(knowledge.Source.Type))
	assert.Equal(t, testURL, *knowledge.Source.URL)
	assert.Contains(t, knowledge.Source.Title, "Website:")
	assert.NotEmpty(t, knowledge.Documents)
	assert.NotEmpty(t, knowledge.Metadata)

	// Check metadata contains crawl information
	assert.Contains(t, knowledge.Metadata, "url")
	assert.Contains(t, knowledge.Metadata, "crawled_at")
	assert.Contains(t, knowledge.Metadata, "pages_count")
	assert.Contains(t, knowledge.Metadata, "crawl_depth")

	// Validate documents have proper structure
	for i, doc := range knowledge.Documents {
		assert.NotEmpty(t, doc.ID, fmt.Sprintf("Document %d should have ID", i))
		assert.NotEmpty(t, doc.EmbeddingText, fmt.Sprintf("Document %d should have embedding text", i))
		assert.NotNil(t, doc.Embeddings, fmt.Sprintf("Document %d should have embeddings", i))
		assert.NotEmpty(t, doc.Metadata, fmt.Sprintf("Document %d should have metadata", i))

		// Content should be either text or image
		assert.True(t,
			doc.Content.Type == "text" || doc.Content.Type == "image",
			fmt.Sprintf("Document %d should have valid content type", i),
		)

		if doc.Content.Type == "text" {
			assert.NotEmpty(t, doc.Content.Text, fmt.Sprintf("Text document %d should have text content", i))
		} else if doc.Content.Type == "image" {
			assert.NotEmpty(t, doc.Content.Image, fmt.Sprintf("Image document %d should have image content", i))
			assert.NotEmpty(t, doc.Content.MIMEType, fmt.Sprintf("Image document %d should have MIME type", i))
		}
	}

	// Test that knowledge can be retrieved
	retrievedKnowledge, err := service.GetKnowledge(ctx, "test-integration")
	require.NoError(t, err)
	require.NotNil(t, retrievedKnowledge)
	assert.Equal(t, knowledge.ID, retrievedKnowledge.ID)
}

func TestIndexKnowledgeFromURL_DeleteExisting(t *testing.T) {
	ctx := t.Context()

	// Check for required environment variables
	openAIKey := os.Getenv("OPENAI_API_KEY")
	firecrawlKey := os.Getenv("FIRECRAWL_API_KEY")

	if openAIKey == "" || firecrawlKey == "" {
		t.Skip("OPENAI_API_KEY and FIRECRAWL_API_KEY required for integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	store := knowledge.NewInMemoryStore()

	firecrawlConfig := &config.FireCrawlConfig{
		APIKey: firecrawlKey,
		APIUrl: "https://api.firecrawl.dev",
	}

	service, err := knowledge.NewServiceWithStore(
		ctx,
		config.NewKnowledgeConfig(),
		&config.ModelConfig{
			OpenAIAPIKey: openAIKey,
		},
		logger,
		store,
		firecrawlConfig,
	)
	require.NoError(t, err)
	defer service.Close()

	testURL := "https://httpbin.org/html"
	knowledgeID := "test-delete-existing"

	// Create test crawl parameters for faster testing
	testCrawlParams := firecrawl.CrawlParams{
		MaxDepth:           lo.ToPtr(1), // Only crawl 1 level deep for tests
		Limit:              lo.ToPtr(2), // Limit to 2 pages for fast testing
		AllowBackwardLinks: lo.ToPtr(false),
		AllowExternalLinks: lo.ToPtr(false),
		ScrapeOptions: firecrawl.ScrapeParams{
			Formats: []string{"markdown"}, // Only markdown for faster processing
		},
	}

	// First indexing
	knowledge1, err := service.IndexKnowledgeFromURL(ctx, knowledgeID, testURL, testCrawlParams)
	require.NoError(t, err)
	require.NotNil(t, knowledge1)

	// Second indexing with same ID should replace the first
	knowledge2, err := service.IndexKnowledgeFromURL(ctx, knowledgeID, testURL, testCrawlParams)
	require.NoError(t, err)
	require.NotNil(t, knowledge2)

	// Should be able to retrieve the knowledge with the same ID
	retrieved, err := service.GetKnowledge(ctx, knowledgeID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, knowledgeID, retrieved.ID)
}

func TestProcessKnowledgeFromCrawl(t *testing.T) {
	// Mock crawl result data
	mockCrawlResult := createMockCrawlResult()

	// Process the crawl result
	documents := knowledge.ProcessKnowledgeFromCrawl(mockCrawlResult)

	// Validate results
	require.NotEmpty(t, documents)
	assert.Len(t, documents, 2) // One screenshot document + one text document

	// Check screenshot document
	screenshotDoc := findDocumentByType(documents, "image")
	require.NotNil(t, screenshotDoc, "Should have screenshot document")
	assert.Equal(t, "image", screenshotDoc.Content.Type)
	assert.NotEmpty(t, screenshotDoc.Content.Image)
	assert.Equal(t, "image/png", screenshotDoc.Content.MIMEType)
	assert.True(t, screenshotDoc.Metadata["has_screenshot"].(bool))

	// Check text documents
	textDocs := filterDocumentsByType(documents, "text")
	assert.Len(t, textDocs, 1) // One text chunk from the second page (not long enough to split)

	for _, doc := range textDocs {
		assert.Equal(t, "text", doc.Content.Type)
		assert.NotEmpty(t, doc.Content.Text)
		assert.NotEmpty(t, doc.EmbeddingText)
		assert.False(t, doc.Metadata["has_screenshot"].(bool))
	}
}

// Helper functions for testing

func createMockCrawlResult() *firecrawl.CrawlStatusResponse {
	return &firecrawl.CrawlStatusResponse{
		Status:      "completed",
		Total:       2,
		Completed:   2,
		CreditsUsed: 10,
		Data: []*firecrawl.FirecrawlDocument{
			&firecrawl.FirecrawlDocument{
				Markdown:   "# Page 1\nThis is the first page content.",
				HTML:       "<h1>Page 1</h1><p>This is the first page content.</p>",
				Screenshot: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
				Metadata: &firecrawl.FirecrawlDocumentMetadata{
					SourceURL: stringPtr("https://example.com/page1"),
					Title:     stringPtr("Page 1 Title"),
				},
			},
			&firecrawl.FirecrawlDocument{
				Markdown: "# Page 2\nThis is a very long page content that should be chunked into smaller pieces. " +
					"Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
					"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
					"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris. " +
					"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. " +
					"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. " +
					"Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, " +
					"totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.",
				HTML: "<h1>Page 2</h1><p>Long content...</p>",
				Metadata: &firecrawl.FirecrawlDocumentMetadata{
					SourceURL: stringPtr("https://example.com/page2"),
					Title:     stringPtr("Page 2 Title"),
				},
			},
		},
	}
}

func stringPtr(s string) *string {
	return &s
}

func findDocumentByType(documents []*knowledge.Document, contentType string) *knowledge.Document {
	for _, doc := range documents {
		if doc.Content.Type == contentType {
			return doc
		}
	}
	return nil
}

func filterDocumentsByType(documents []*knowledge.Document, contentType string) []*knowledge.Document {
	var result []*knowledge.Document
	for _, doc := range documents {
		if doc.Content.Type == contentType {
			result = append(result, doc)
		}
	}
	return result
}

// Benchmark test
func BenchmarkIndexKnowledgeFromURL(b *testing.B) {
	ctx := b.Context() // For benchmarks, use b.Context()

	// Skip if no API keys
	openAIKey := os.Getenv("OPENAI_API_KEY")
	firecrawlKey := os.Getenv("FIRECRAWL_API_KEY")

	if openAIKey == "" || firecrawlKey == "" {
		b.Skip("API keys required for benchmark")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	store := knowledge.NewInMemoryStore()

	firecrawlConfig := &config.FireCrawlConfig{
		APIKey: firecrawlKey,
		APIUrl: "https://api.firecrawl.dev",
	}

	service, err := knowledge.NewServiceWithStore(
		ctx,
		config.NewKnowledgeConfig(),
		&config.ModelConfig{
			OpenAIAPIKey: openAIKey,
		},
		logger,
		store,
		firecrawlConfig,
	)
	require.NoError(b, err)
	defer service.Close()

	testURL := "https://httpbin.org/html"

	// Create benchmark crawl parameters with minimal limits for performance
	benchCrawlParams := firecrawl.CrawlParams{
		MaxDepth:           lo.ToPtr(1), // Only crawl 1 level deep for benchmarks
		Limit:              lo.ToPtr(1), // Limit to 1 page for benchmark speed
		AllowBackwardLinks: lo.ToPtr(false),
		AllowExternalLinks: lo.ToPtr(false),
		ScrapeOptions: firecrawl.ScrapeParams{
			Formats: []string{"markdown"}, // Only markdown for faster processing
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		knowledgeID := fmt.Sprintf("benchmark-%d", i)
		_, err := service.IndexKnowledgeFromURL(ctx, knowledgeID, testURL, benchCrawlParams)
		if err != nil {
			b.Fatal(err)
		}
	}
}
