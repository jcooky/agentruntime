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

func TestConversationSummarizer_CountTokens(t *testing.T) {
	ctx := context.Background()

	logger := mylog.NewLogger("debug", "text")
	modelConfig := &config.ModelConfig{
		AnthropicAPIKey: "dummy-key-for-testing",
	}
	g, err := genkit.NewGenkit(ctx, modelConfig, logger, false)
	require.NoError(t, err)

	testConfig := &config.ModelConfig{
		AnthropicAPIKey: "dummy-key-for-testing",
		ConversationSummary: config.ConversationSummaryConfig{
			MaxTokens:                   1000,
			SummaryTokens:               200,
			MinConversationsToSummarize: 5,
			ModelForSummary:             "openai/gpt-4o-mini",
		},
	}

	summarizer, err := NewConversationSummarizer(g, testConfig)
	require.NoError(t, err)

	// Test token counting
	testCases := []struct {
		name      string
		text      string
		minTokens int
		maxTokens int
	}{
		{
			name:      "empty string",
			text:      "",
			minTokens: 0,
			maxTokens: 0,
		},
		{
			name:      "simple text",
			text:      "Hello, world!",
			minTokens: 1,
			maxTokens: 10,
		},
		{
			name:      "longer text",
			text:      "This is a longer sentence with more words to test token counting functionality.",
			minTokens: 10,
			maxTokens: 25,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := summarizer.CountTokens(tc.text)
			assert.GreaterOrEqual(t, tokens, tc.minTokens)
			assert.LessOrEqual(t, tokens, tc.maxTokens)
		})
	}
}

func TestConversationSummarizer_ProcessConversationHistory(t *testing.T) {
	ctx := context.Background()

	logger := mylog.NewLogger("debug", "text")
	modelConfig := &config.ModelConfig{
		AnthropicAPIKey: "dummy-key-for-testing",
	}
	g, err := genkit.NewGenkit(ctx, modelConfig, logger, false)
	require.NoError(t, err)

	testConfig := &config.ModelConfig{
		AnthropicAPIKey: "dummy-key-for-testing",
		ConversationSummary: config.ConversationSummaryConfig{
			MaxTokens:                   500, // Low limit to force summarization
			SummaryTokens:               100,
			MinConversationsToSummarize: 3,
			ModelForSummary:             "openai/gpt-4o-mini",
		},
	}

	summarizer, err := NewConversationSummarizer(g, testConfig)
	require.NoError(t, err)

	t.Run("empty conversations", func(t *testing.T) {
		result, err := summarizer.ProcessConversationHistory(ctx, []Conversation{}, []File{})
		require.NoError(t, err)

		assert.Empty(t, result.RecentConversations)
		assert.Nil(t, result.Summary)
		assert.Equal(t, 0, result.TotalTokens)
		assert.False(t, result.WasSummarizationUsed)
	})

	t.Run("conversations under token limit", func(t *testing.T) {
		conversations := []Conversation{
			{User: "user1", Text: "Hello"},
			{User: "bot", Text: "Hi there!"},
		}

		result, err := summarizer.ProcessConversationHistory(ctx, conversations, []File{})
		require.NoError(t, err)

		assert.Equal(t, conversations, result.RecentConversations)
		assert.Nil(t, result.Summary)
		assert.False(t, result.WasSummarizationUsed)
	})

	t.Run("many conversations under min threshold", func(t *testing.T) {
		// Create conversations that exceed token limit but are under min count
		conversations := []Conversation{
			{User: "user1", Text: "This is a very long message that contains many words and should consume a significant number of tokens to test the token counting and truncation functionality of the conversation summarizer system."},
			{User: "bot", Text: "This is another very long response that also contains many words and should consume a significant number of tokens to test the token counting and truncation functionality."},
		}

		result, err := summarizer.ProcessConversationHistory(ctx, conversations, []File{})
		require.NoError(t, err)

		// Should truncate without summarizing since below min threshold
		assert.LessOrEqual(t, len(result.RecentConversations), len(conversations))
		assert.Nil(t, result.Summary)
		assert.False(t, result.WasSummarizationUsed)
	})

	t.Run("conversations with request files", func(t *testing.T) {
		conversations := []Conversation{
			{User: "user1", Text: "Hello"},
			{User: "bot", Text: "Hi there!"},
		}

		// Add a file to the request
		requestFiles := []File{
			{
				ContentType: "image/jpeg",
				Data:        base64.StdEncoding.EncodeToString(make([]byte, 50*1024)), // 50KB
				Filename:    "test.jpg",
			},
		}

		result, err := summarizer.ProcessConversationHistory(ctx, conversations, requestFiles)
		require.NoError(t, err)

		// Should include file tokens in total
		conversationOnlyResult, _ := summarizer.ProcessConversationHistory(ctx, conversations, []File{})
		assert.Greater(t, result.TotalTokens, conversationOnlyResult.TotalTokens)
		assert.Equal(t, conversations, result.RecentConversations)
		assert.Nil(t, result.Summary)
		assert.False(t, result.WasSummarizationUsed)
	})
}

func TestConversationSummarizer_findSplitPoint(t *testing.T) {
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

	conversations := make([]Conversation, 20)
	for i := 0; i < 20; i++ {
		conversations[i] = Conversation{
			User: "user1",
			Text: "Short message",
		}
	}

	splitPoint := summarizer.findSplitPoint(conversations)

	// Should find a valid split point
	assert.Greater(t, splitPoint, 0)
	assert.Less(t, splitPoint, len(conversations))

	// Remaining conversations should be at least the minimum
	remaining := len(conversations) - splitPoint
	minRequired := len(conversations) / 3
	if minRequired < testConfig.ConversationSummary.MinConversationsToSummarize {
		minRequired = testConfig.ConversationSummary.MinConversationsToSummarize
	}
	assert.GreaterOrEqual(t, remaining, minRequired)
}

func TestConversationSummarizer_truncateToTokenLimit(t *testing.T) {
	ctx := context.Background()

	logger := mylog.NewLogger("debug", "text")
	testConfig := &config.ModelConfig{
		AnthropicAPIKey: "dummy-key-for-testing",
		ConversationSummary: config.ConversationSummaryConfig{
			MaxTokens:                   100,
			SummaryTokens:               50,
			MinConversationsToSummarize: 3,
			ModelForSummary:             "openai/gpt-4o-mini",
		},
	}
	g, err := genkit.NewGenkit(ctx, testConfig, logger, false)
	require.NoError(t, err)

	summarizer, err := NewConversationSummarizer(g, testConfig)
	require.NoError(t, err)

	conversations := []Conversation{
		{User: "user1", Text: "First message with some content"},
		{User: "user2", Text: "Second message with some content"},
		{User: "user3", Text: "Third message with some content"},
		{User: "user4", Text: "Fourth message with some content"},
		{User: "user5", Text: "Fifth message with some content"},
	}

	result := summarizer.truncateToTokenLimit(conversations, 50)

	// Should return some conversations (from the end)
	assert.Greater(t, len(result), 0)
	assert.LessOrEqual(t, len(result), len(conversations))

	// Should not exceed token limit
	tokens := summarizer.CountConversationTokens(result)
	assert.LessOrEqual(t, tokens, 50)

	// Should preserve order and keep recent conversations
	if len(result) > 0 {
		assert.Equal(t, conversations[len(conversations)-1], result[len(result)-1])
	}
}
