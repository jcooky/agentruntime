package engine

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/genkit"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageTokenCalculator_CalculateImageTokens(t *testing.T) {
	calc := NewImageTokenCalculator()

	testCases := []struct {
		name        string
		contentType string
		dataSize    int
		expectedMin int
		expectedMax int
	}{
		{
			name:        "small JPEG",
			contentType: "image/jpeg",
			dataSize:    50 * 1024, // 50KB
			expectedMin: 85,        // At least base tokens
			expectedMax: 500,       // Conservative upper bound
		},
		{
			name:        "medium PNG",
			contentType: "image/png",
			dataSize:    500 * 1024, // 500KB
			expectedMin: 85,
			expectedMax: 1000,
		},
		{
			name:        "large image",
			contentType: "image/jpeg",
			dataSize:    2 * 1024 * 1024, // 2MB
			expectedMin: 200,
			expectedMax: 2000,
		},
		{
			name:        "PDF file",
			contentType: "application/pdf",
			dataSize:    1024 * 1024, // 1MB
			expectedMin: 100,
			expectedMax: 300000, // PDF can have lots of text
		},
		{
			name:        "text file",
			contentType: "text/plain",
			dataSize:    10000, // 10KB
			expectedMin: 1000,
			expectedMax: 10000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create dummy base64 data of specified size
			dummyData := make([]byte, tc.dataSize)
			for i := range dummyData {
				dummyData[i] = byte(i % 256)
			}
			base64Data := base64.StdEncoding.EncodeToString(dummyData)

			tokens := calc.CalculateImageTokens(tc.contentType, base64Data)
			assert.GreaterOrEqual(t, tokens, tc.expectedMin)
			assert.LessOrEqual(t, tokens, tc.expectedMax)
		})
	}
}

func TestImageTokenCalculator_calculateOpenAIImageTokens(t *testing.T) {
	calc := NewImageTokenCalculator()

	testCases := []struct {
		name     string
		width    int
		height   int
		expected int
	}{
		{
			name:   "small square image",
			width:  512,
			height: 512,
			// After scaling to shortest side 768px: 768x768 = 2x2 tiles
			expected: 85 + (2 * 2 * 170), // Base + 4 tiles
		},
		{
			name:   "medium rectangle",
			width:  1024,
			height: 768,
			// After scaling to shortest side 768px: 1024x768 = 2x2 tiles
			expected: 85 + (2 * 2 * 170), // Base + 4 tiles
		},
		{
			name:   "large image",
			width:  2048,
			height: 1536,
			// After scaling to shortest side 768px: 1024x768 = 2x2 tiles
			expected: 85 + (2 * 2 * 170), // Base + 4 tiles
		},
		{
			name:   "very small image",
			width:  200,
			height: 200,
			// After scaling to shortest side 768px: 768x768 = 2x2 tiles
			expected: 85 + (2 * 2 * 170), // Base + 4 tiles
		},
		{
			name:   "wide panorama",
			width:  1536,
			height: 512,
			// After scaling to shortest side 768px: 2304x768 = 5x2 tiles
			expected: 85 + (5 * 2 * 170), // Base + 10 tiles
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := calc.calculateOpenAIImageTokens(tc.width, tc.height)

			// Allow some tolerance for rounding in calculations
			tolerance := 170 // One tile worth of tokens
			assert.InDelta(t, tc.expected, tokens, float64(tolerance))
		})
	}
}

func TestImageTokenCalculator_scaleToFit(t *testing.T) {
	calc := NewImageTokenCalculator()

	testCases := []struct {
		name                          string
		width, height                 int
		maxWidth, maxHeight           int
		expectedWidth, expectedHeight int
	}{
		{
			name:  "no scaling needed",
			width: 800, height: 600,
			maxWidth: 1024, maxHeight: 1024,
			expectedWidth: 800, expectedHeight: 600,
		},
		{
			name:  "scale down proportionally",
			width: 3000, height: 2000,
			maxWidth: 1500, maxHeight: 1500,
			expectedWidth: 1500, expectedHeight: 1000,
		},
		{
			name:  "scale down tall image",
			width: 1000, height: 3000,
			maxWidth: 1024, maxHeight: 1024,
			expectedWidth: 341, expectedHeight: 1024, // Approximately
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualWidth, actualHeight := calc.scaleToFit(tc.width, tc.height, tc.maxWidth, tc.maxHeight)

			// Allow some rounding tolerance
			assert.InDelta(t, tc.expectedWidth, actualWidth, 5)
			assert.InDelta(t, tc.expectedHeight, actualHeight, 5)

			// Verify that scaled dimensions don't exceed maximums
			assert.LessOrEqual(t, actualWidth, tc.maxWidth)
			assert.LessOrEqual(t, actualHeight, tc.maxHeight)
		})
	}
}

func TestConversationSummarizer_CountFileTokens(t *testing.T) {
	ctx := context.Background()

	logger := mylog.NewLogger("debug", "text")
	testConfig := &config.ModelConfig{
		AnthropicAPIKey: "dummy-key-for-testing",
		ConversationSummary: config.ConversationSummaryConfig{
			MaxTokens:                   1000,
			SummaryTokens:               200,
			MinConversationsToSummarize: 5,
			ModelForSummary:             "openai/gpt-4o-mini",
		},
	}
	g, err := genkit.NewGenkit(ctx, testConfig, logger, false)
	require.NoError(t, err)

	summarizer, err := NewConversationSummarizer(g, testConfig)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		contentType string
		dataSize    int
		minTokens   int
		maxTokens   int
	}{
		{
			name:        "JPEG image",
			contentType: "image/jpeg",
			dataSize:    100 * 1024, // 100KB
			minTokens:   85,
			maxTokens:   1000,
		},
		{
			name:        "PNG image",
			contentType: "image/png",
			dataSize:    50 * 1024, // 50KB
			minTokens:   85,
			maxTokens:   500,
		},
		{
			name:        "PDF document",
			contentType: "application/pdf",
			dataSize:    500 * 1024, // 500KB
			minTokens:   50000,      // PDFs can have lots of text
			maxTokens:   200000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create dummy base64 data
			dummyData := make([]byte, tc.dataSize)
			base64Data := base64.StdEncoding.EncodeToString(dummyData)

			tokens := summarizer.CountFileTokens(tc.contentType, base64Data)
			assert.GreaterOrEqual(t, tokens, tc.minTokens)
			assert.LessOrEqual(t, tokens, tc.maxTokens)
		})
	}
}

func TestConversationSummarizer_CountRequestFilesTokens(t *testing.T) {
	ctx := context.Background()

	logger := mylog.NewLogger("debug", "text")
	testConfig := &config.ModelConfig{
		AnthropicAPIKey: "dummy-key-for-testing",
		ConversationSummary: config.ConversationSummaryConfig{
			MaxTokens:                   5000,
			SummaryTokens:               200,
			MinConversationsToSummarize: 3,
			ModelForSummary:             "openai/gpt-4o-mini",
		},
	}
	g, err := genkit.NewGenkit(ctx, testConfig, logger, false)
	require.NoError(t, err)

	summarizer, err := NewConversationSummarizer(g, testConfig)
	require.NoError(t, err)

	// Test request files token counting
	requestFiles := []File{
		{
			ContentType: "image/jpeg",
			Data:        base64.StdEncoding.EncodeToString(make([]byte, 50*1024)), // 50KB image
			Filename:    "test.jpg",
		},
	}

	// Count tokens for request files
	fileTokens := summarizer.CountRequestFilesTokens(requestFiles)
	assert.Greater(t, fileTokens, 0)
	assert.Greater(t, fileTokens, 50) // Should be at least 50 tokens for image

	// Test with no files
	noFileTokens := summarizer.CountRequestFilesTokens([]File{})
	assert.Equal(t, 0, noFileTokens)
}
