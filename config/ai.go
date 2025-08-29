package config

type (
	// ConversationSummaryConfig holds configuration for conversation summarization
	ConversationSummaryConfig struct {
		// MaxTokens is the maximum total tokens allowed in conversation history
		MaxTokens int `json:"max_tokens"`
		// SummaryTokens is the target token count for each summary
		SummaryTokens int `json:"summary_tokens"`
		// MinConversationsToSummarize is the minimum number of conversations to trigger summarization
		MinConversationsToSummarize int `json:"min_conversations_to_summarize"`
		// ModelForSummary is the model to use for generating summaries
		ModelForSummary string `json:"model_for_summary"`
		// TokenProvider specifies which token counter to use ("openai", "anthropic", "auto")
		// "auto" will automatically detect based on ModelForSummary
		TokenProvider string `json:"token_provider"`
	}

	ModelConfig struct {
		OpenAIAPIKey        string                    `json:"openaiApiKey"`
		XAIAPIKey           string                    `json:"xaiApiKey"`
		AnthropicAPIKey     string                    `json:"anthropicApiKey"`
		TraceVerbose        bool                      `json:"traceVerbose"`
		ConversationSummary ConversationSummaryConfig `json:"conversationSummary"`
	}
)

// DefaultConversationSummaryConfig returns the default configuration
// Always uses Anthropic API for token counting
func DefaultConversationSummaryConfig() ConversationSummaryConfig {
	return ConversationSummaryConfig{
		MaxTokens:                   100000,              // 100k tokens limit
		SummaryTokens:               2000,                // 2k tokens per summary
		MinConversationsToSummarize: 10,                  // At least 10 conversations before summarizing
		ModelForSummary:             "openai/gpt-5-mini", // Use efficient model for summaries
		TokenProvider:               "anthropic",         // Always use Anthropic API for token counting
	}
}
