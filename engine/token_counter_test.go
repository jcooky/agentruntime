package engine

import (
	"context"
	"testing"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/genkit"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenCounterFactory(t *testing.T) {
	modelConfig := &config.ModelConfig{
		AnthropicAPIKey: "dummy-key-for-testing",
	}
	factory := NewDefaultTokenCounterFactory(modelConfig)

	t.Run("create OpenAI token counter", func(t *testing.T) {
		counter, err := factory.CreateTokenCounter("openai")
		require.NoError(t, err)
		assert.Equal(t, "openai", counter.ProviderName())
		assert.IsType(t, &OpenAITokenCounter{}, counter)
	})

	t.Run("create Anthropic token counter", func(t *testing.T) {
		counter, err := factory.CreateTokenCounter("anthropic")
		require.NoError(t, err)
		assert.Equal(t, "anthropic", counter.ProviderName())
		assert.IsType(t, &AnthropicTokenCounter{}, counter)
	})

	t.Run("create token counter for model", func(t *testing.T) {
		// GPT models should use OpenAI counter
		counter, err := factory.CreateTokenCounterForModel("openai/gpt-4")
		require.NoError(t, err)
		assert.Equal(t, "openai", counter.ProviderName())

		// Claude models should use Anthropic counter
		counter, err = factory.CreateTokenCounterForModel("anthropic/claude-3-5-sonnet")
		require.NoError(t, err)
		assert.Equal(t, "anthropic", counter.ProviderName())
	})

	t.Run("unsupported provider", func(t *testing.T) {
		_, err := factory.CreateTokenCounter("unsupported")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported token counter provider")
	})
}

func TestOpenAITokenCounter(t *testing.T) {
	counter, err := NewOpenAITokenCounter()
	require.NoError(t, err)

	t.Run("count simple text tokens", func(t *testing.T) {
		tokens := counter.CountTokens("Hello world")
		assert.Greater(t, tokens, 0)
		assert.Less(t, tokens, 10) // Should be around 2-3 tokens
	})

	t.Run("count empty text", func(t *testing.T) {
		tokens := counter.CountTokens("")
		assert.Equal(t, 0, tokens)
	})

	t.Run("count conversation tokens", func(t *testing.T) {
		conversations := []Conversation{
			{User: "user", Text: "Hello"},
			{User: "assistant", Text: "Hi there!"},
		}
		tokens := counter.CountConversationTokens(conversations)
		assert.Greater(t, tokens, 0)
	})

	t.Run("provider name", func(t *testing.T) {
		assert.Equal(t, "openai", counter.ProviderName())
	})
}

func TestAnthropicTokenCounter(t *testing.T) {
	// Use dummy API key for testing
	apiKey := "dummy-key-for-testing"

	counter, err := NewAnthropicTokenCounter(apiKey, "claude-3-5-sonnet-20241022")
	require.NoError(t, err)

	t.Run("count simple text tokens", func(t *testing.T) {
		tokens := counter.CountTokens("Hello world")
		assert.Greater(t, tokens, 0)
		// Note: This would actually call the API in a real test
	})

	t.Run("provider name", func(t *testing.T) {
		assert.Equal(t, "anthropic", counter.ProviderName())
	})
}

func TestConversationSummarizerWithDifferentTokenCounters(t *testing.T) {
	ctx := context.Background()
	logger := mylog.NewLogger("debug", "text")
	modelConfig := &config.ModelConfig{
		AnthropicAPIKey: "dummy-key-for-testing",
	}
	g, err := genkit.NewGenkit(ctx, modelConfig, logger, false)
	require.NoError(t, err)

	t.Run("OpenAI token counter", func(t *testing.T) {
		testConfig := &config.ModelConfig{
			AnthropicAPIKey: "dummy-key-for-testing",
			ConversationSummary: config.ConversationSummaryConfig{
				MaxTokens:                   5000,
				SummaryTokens:               200,
				MinConversationsToSummarize: 3,
				ModelForSummary:             "openai/gpt-4o-mini",
				TokenProvider:               "openai",
			},
		}

		summarizer, err := NewConversationSummarizer(g, testConfig)
		require.NoError(t, err)
		assert.Equal(t, "openai", summarizer.GetTokenCounter().ProviderName())

		// Test basic functionality
		tokens := summarizer.CountTokens("Hello world")
		assert.Greater(t, tokens, 0)
	})

	t.Run("Anthropic token counter", func(t *testing.T) {
		testConfig := &config.ModelConfig{
			AnthropicAPIKey: "dummy-key-for-testing",
			ConversationSummary: config.ConversationSummaryConfig{
				MaxTokens:                   5000,
				SummaryTokens:               200,
				MinConversationsToSummarize: 3,
				ModelForSummary:             "anthropic/claude-3-5-sonnet",
				TokenProvider:               "anthropic",
			},
		}

		summarizer, err := NewConversationSummarizer(g, testConfig)
		require.NoError(t, err)
		assert.Equal(t, "anthropic", summarizer.GetTokenCounter().ProviderName())
	})

	t.Run("auto-detect from model name", func(t *testing.T) {
		// Should use OpenAI counter for GPT models
		testConfig := &config.ModelConfig{
			AnthropicAPIKey: "dummy-key-for-testing",
			ConversationSummary: config.ConversationSummaryConfig{
				MaxTokens:                   5000,
				SummaryTokens:               200,
				MinConversationsToSummarize: 3,
				ModelForSummary:             "openai/gpt-4o-mini",
				TokenProvider:               "auto",
			},
		}

		summarizer, err := NewConversationSummarizer(g, testConfig)
		require.NoError(t, err)
		assert.Equal(t, "openai", summarizer.GetTokenCounter().ProviderName())

		// Should use Anthropic counter for Claude models
		testConfig.ConversationSummary.ModelForSummary = "anthropic/claude-3-5-sonnet"
		summarizer, err = NewConversationSummarizer(g, testConfig)
		require.NoError(t, err)
		assert.Equal(t, "anthropic", summarizer.GetTokenCounter().ProviderName())
	})
}

func TestGetProviderFromModel(t *testing.T) {
	testCases := []struct {
		model    string
		expected string
	}{
		{"gpt-4", "openai"},
		{"gpt-3.5-turbo", "openai"},
		{"claude-3-5-sonnet", "anthropic"},
		{"claude-opus", "anthropic"},
		{"grok-1", "xai"},
		{"unknown-model", "openai"}, // fallback
	}

	for _, tc := range testCases {
		t.Run(tc.model, func(t *testing.T) {
			provider := GetProviderFromModel(tc.model)
			assert.Equal(t, tc.expected, provider)
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	t.Run("estimate text tokens", func(t *testing.T) {
		tokens := EstimateTokens("Hello world")
		assert.Greater(t, tokens, 0)
		assert.Equal(t, 2, tokens) // "Hello world" = 11 chars / 4 = 2.75 -> 2
	})

	t.Run("estimate empty text", func(t *testing.T) {
		tokens := EstimateTokens("")
		assert.Equal(t, 0, tokens)
	})
}

func TestEstimateFileTokens(t *testing.T) {
	testCases := []struct {
		contentType string
		dataSize    int
		description string
	}{
		{"image/jpeg", 1000, "JPEG image"},
		{"image/png", 1000, "PNG image"},
		{"application/pdf", 1000, "PDF document"},
		{"text/plain", 1000, "text file"},
		{"application/octet-stream", 1000, "binary file"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			data := make([]byte, tc.dataSize)
			tokens := EstimateFileTokens(tc.contentType, string(data))
			assert.Greater(t, tokens, 0)
		})
	}
}
