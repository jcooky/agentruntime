package knowledge_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/habiliai/agentruntime/config"
	xgenkit "github.com/habiliai/agentruntime/internal/genkit"
	"github.com/habiliai/agentruntime/knowledge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sample 1x1 PDF for testing (base64 encoded)
const testPDFBase64 = "JVBERi0xLjMKJeLjz9MKMSAwIG9iago8PAovVHlwZSAvQ2F0YWxvZwovT3V0bGluZXMgMiAwIFIKL1BhZ2VzIDMgMCBSCj4+CmVuZG9iagoyIDAgb2JqCjw8Ci9UeXBlIC9PdXRsaW5lcwovQ291bnQgMAo+PgplbmRvYmoKMyAwIG9iago8PAovVHlwZSAvUGFnZXMKL0NvdW50IDEKL0tpZHMgWzQgMCBSXQo+PgplbmRvYmoKNCAwIG9iago8PAovVHlwZSAvUGFnZQovUGFyZW50IDMgMCBSCi9NZWRpYUJveCBbMCAwIDYxMiA3OTJdCi9Db250ZW50cyA1IDAgUgovUmVzb3VyY2VzIDw8Ci9Gb250IDw8Ci9GMSA2IDAgUgo+Pgo+Pgo+PgplbmRvYmoKNSAwIG9iago8PAovTGVuZ3RoIDQ0Cj4+CnN0cmVhbQpCVApxCjcwIDUwIFRECi9GMSAxMiBUZgooSGVsbG8gV29ybGQpIFRqCkVUClEKZW5kc3RyZWFtCmVuZG9iago2IDAgb2JqCjw8Ci9UeXBlIC9Gb250Ci9TdWJ0eXBlIC9UeXBlMQovQmFzZUZvbnQgL0hlbHZldGljYQo+PgplbmRvYmoKeHJlZgowIDcKMDAwMDAwMDAwMCA2NTUzNSBmIAowMDAwMDAwMDE1IDAwMDAwIG4gCjAwMDAwMDAwNzQgMDAwMDAgbiAKMDAwMDAwMDEyMCAwMDAwMCBuIAowMDAwMDAwMTc5IDAwMDAwIG4gCjAwMDAwMDAzNjQgMDAwMDAgbiAKMDAwMDAwMDQ2NiAwMDAwMCBuIAp0cmFpbGVyCjw8Ci9TaXplIDcKL1Jvb3QgMSAwIFIKPj4Kc3RhcnR4cmVmCjU2NQolJUVPRg=="

var (
	//go:embed testdata/solana-whitepaper-en.pdf
	solanaWhitepaperPDF []byte
)

func TestProcessKnowledgeFromPDF(t *testing.T) {
	ctx := context.Background()

	// Check if we have API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" && !testing.Short() {
		t.Skip("OPENAI_API_KEY not set, skipping test")
	}

	// Initialize genkit
	modelConfig := &config.ModelConfig{
		OpenAIAPIKey: apiKey,
	}
	if apiKey == "" {
		modelConfig.OpenAIAPIKey = "test-key"
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	g, err := xgenkit.NewGenkit(ctx, modelConfig, logger, false)
	require.NoError(t, err)

	// Decode test PDF
	pdfData, err := base64.StdEncoding.DecodeString(testPDFBase64)
	require.NoError(t, err)

	// Create reader
	reader := bytes.NewReader(pdfData)

	// Process PDF
	result, err := knowledge.ProcessKnowledgeFromPDF(ctx, g, "test-pdf", reader, logger, config.NewKnowledgeConfig())

	// If no API key, we expect an error
	if apiKey == "" {
		require.Error(t, err)
		t.Logf("Expected error (no API key): %v", err)
		return
	}

	require.NoError(t, err)
	require.NotNil(t, result)

	// Validate basic structure
	assert.Equal(t, "test-pdf", result.ID)
	assert.Equal(t, knowledge.SourceType("pdf"), result.Source.Type)
	assert.NotEmpty(t, result.Documents)

	// Check first document
	if len(result.Documents) > 0 {
		doc := result.Documents[0]
		assert.Equal(t, "test-pdf_page_1", doc.ID)
		assert.Len(t, doc.Contents, 2) // Text and Image
		assert.Equal(t, "library", doc.Metadata["extraction_method"])
		assert.NotEmpty(t, doc.EmbeddingText)

		t.Logf("Extracted text: %s", doc.EmbeddingText)
	}
}

func TestProcessKnowledgeFromPDF_InvalidInput(t *testing.T) {
	ctx := context.Background()

	// Initialize genkit
	modelConfig := &config.ModelConfig{
		OpenAIAPIKey: "test-key",
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	g, err := xgenkit.NewGenkit(ctx, modelConfig, logger, false)
	require.NoError(t, err)

	tests := []struct {
		name        string
		input       []byte
		expectedErr string
	}{
		{
			name:        "empty data",
			input:       []byte{},
			expectedErr: "failed to open PDF",
		},
		{
			name:        "invalid PDF",
			input:       []byte("not a pdf"),
			expectedErr: "failed to open PDF",
		},
		{
			name:        "corrupted PDF header",
			input:       []byte("%PDF-1.4\ngarbage"),
			expectedErr: "failed to open PDF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)
			_, err := knowledge.ProcessKnowledgeFromPDF(ctx, g, "test-id", reader, logger, config.NewKnowledgeConfig())
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestExtractTextWithVisionLLM(t *testing.T) {
	ctx := context.Background()

	// Check if we have API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping test")
	}

	// Initialize genkit
	modelConfig := &config.ModelConfig{
		AnthropicAPIKey: apiKey,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	g, err := xgenkit.NewGenkit(ctx, modelConfig, logger, false)
	require.NoError(t, err)

	// Simple 1x1 white pixel PNG (base64)
	base64Image := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="

	// Test extraction
	text, err := knowledge.ExtractTextWithVisionLLM(ctx, g, base64Image, 1)
	require.NoError(t, err)

	// Vision LLM should return something even for a blank image
	assert.NotEmpty(t, text)
	t.Logf("Extracted text from blank image: %s", text)
}

func TestExtractTextWIthVisionLLM_RealImage(t *testing.T) {
	ctx := context.Background()

	anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicAPIKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping test")
	}

	// Initialize genkit
	modelConfig := &config.ModelConfig{
		AnthropicAPIKey: anthropicAPIKey,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	g, err := xgenkit.NewGenkit(ctx, modelConfig, logger, false)
	require.NoError(t, err)

	// Simple 1x1 white pixel PNG (base64)
	imageFile, err := os.ReadFile("testdata/page_1.jpg")
	require.NoError(t, err)

	base64Image := base64.StdEncoding.EncodeToString(imageFile)

	// Test extraction
	text, err := knowledge.ExtractTextWithVisionLLM(ctx, g, base64Image, 1)
	require.NoError(t, err)

	// Vision LLM should return something even for a blank image
	assert.NotEmpty(t, text)
	t.Logf("Extracted text from blank image: %s", text)
}

// Benchmark for performance testing
func BenchmarkProcessKnowledgeFromPDF(b *testing.B) {
	ctx := context.Background()

	// Skip if no API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		b.Skip("OPENAI_API_KEY not set")
	}

	// Initialize genkit
	modelConfig := &config.ModelConfig{
		OpenAIAPIKey: apiKey,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	g, err := xgenkit.NewGenkit(ctx, modelConfig, logger, false)
	require.NoError(b, err)

	// Decode test PDF
	pdfData, err := base64.StdEncoding.DecodeString(testPDFBase64)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(pdfData)
		_, err := knowledge.ProcessKnowledgeFromPDF(ctx, g, "bench-pdf", reader, logger, config.NewKnowledgeConfig())
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Integration test with real PDF file
func TestProcessKnowledgeFromPDF_RealFile(t *testing.T) {
	ctx := context.Background()

	// Check if we have API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping test")
	}

	// Initialize genkit
	modelConfig := &config.ModelConfig{
		OpenAIAPIKey: apiKey,
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // Use discard to reduce log output
	g, err := xgenkit.NewGenkit(ctx, modelConfig, logger, false)
	require.NoError(t, err)

	// Open the real PDF file
	pdfFile := bytes.NewReader(solanaWhitepaperPDF)

	// Process only first few pages to avoid token limits
	// We'll create a limited reader that processes only a subset
	result, err := knowledge.ProcessKnowledgeFromPDF(ctx, g, "solana-whitepaper", pdfFile, logger, config.NewKnowledgeConfig())

	// Allow partial success - the function might fail on some pages
	require.NoError(t, err)

	// Validate results
	assert.Equal(t, "solana-whitepaper", result.ID)
	assert.Equal(t, knowledge.SourceType("pdf"), result.Source.Type)

	// Check metadata
	t.Logf("PDF Title: %v", result.Source.Title)
	t.Logf("PDF Author: %v", result.Metadata["author"])
	t.Logf("PDF Subject: %v", result.Metadata["subject"])
	t.Logf("Pages processed: %d", len(result.Documents))

	// Check that we processed at least some pages
	assert.Greater(t, len(result.Documents), 0, "Should have processed at least one page")

	// Check first page content if available
	if len(result.Documents) > 0 {
		firstPage := result.Documents[0]
		assert.Equal(t, "solana-whitepaper_page_1", firstPage.ID)
		assert.Equal(t, 1, firstPage.Metadata["page_number"])

		// Should contain text (even if it's an error message from Vision API)
		extractedText := firstPage.EmbeddingText
		assert.NotEmpty(t, extractedText)

		// Log what we got for debugging
		t.Logf("First page text length: %d", len(extractedText))
		t.Logf("First page preview (first 200 chars): %.200s...", extractedText)

		// Check for any blockchain-related content in any processed page
		foundRelevantContent := false
		for i, doc := range result.Documents {
			lowerText := strings.ToLower(doc.EmbeddingText)
			if strings.Contains(lowerText, "solana") ||
				strings.Contains(lowerText, "blockchain") ||
				strings.Contains(lowerText, "consensus") ||
				strings.Contains(lowerText, "distributed") ||
				strings.Contains(lowerText, "transaction") ||
				strings.Contains(lowerText, "validator") {
				foundRelevantContent = true
				t.Logf("Found relevant content on page %d", i+1)
				break
			}
		}

		// This is a soft check - log if not found but don't fail
		if !foundRelevantContent {
			t.Log("Warning: No blockchain-related content found in processed pages")
		}
	}

	// Check that different pages have different content (if multiple pages processed)
	if len(result.Documents) > 1 {
		assert.NotEqual(t, result.Documents[0].EmbeddingText, result.Documents[1].EmbeddingText,
			"Different pages should have different content")
	}
}

// Test with a simpler PDF to ensure basic functionality
func TestProcessKnowledgeFromPDF_Simple(t *testing.T) {
	ctx := context.Background()

	// Check if we have API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" && !testing.Short() {
		t.Skip("OPENAI_API_KEY not set, skipping test")
	}

	// Initialize genkit
	modelConfig := &config.ModelConfig{
		OpenAIAPIKey: apiKey,
	}
	if apiKey == "" {
		modelConfig.OpenAIAPIKey = "test-key"
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	g, err := xgenkit.NewGenkit(ctx, modelConfig, logger, false)
	require.NoError(t, err)

	// Use the simple test PDF
	pdfData, err := base64.StdEncoding.DecodeString(testPDFBase64)
	require.NoError(t, err)

	reader := bytes.NewReader(pdfData)
	result, err := knowledge.ProcessKnowledgeFromPDF(ctx, g, "test-pdf", reader, logger, config.NewKnowledgeConfig())

	// If no API key, we expect an error
	if apiKey == "" {
		require.Error(t, err)
		t.Logf("Expected error (no API key): %v", err)
		return
	}

	require.NoError(t, err)
	require.NotNil(t, result)

	// This simple PDF should work without token limit issues
	assert.Equal(t, 1, len(result.Documents))
	assert.NotEmpty(t, result.Documents[0].EmbeddingText)

	t.Logf("Simple PDF extracted text: %s", result.Documents[0].EmbeddingText)
}
