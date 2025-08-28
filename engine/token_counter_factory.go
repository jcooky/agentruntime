package engine

import (
	"fmt"
	"strings"

	"github.com/habiliai/agentruntime/config"
)

// DefaultTokenCounterFactory is the default implementation of TokenCounterFactory
type DefaultTokenCounterFactory struct {
	anthropicApiKey string
}

// NewDefaultTokenCounterFactory creates a new default token counter factory
func NewDefaultTokenCounterFactory(config *config.ModelConfig) *DefaultTokenCounterFactory {
	return &DefaultTokenCounterFactory{
		anthropicApiKey: config.AnthropicAPIKey,
	}
}

// CreateTokenCounter creates a token counter for the specified provider
func (f *DefaultTokenCounterFactory) CreateTokenCounter(provider string) (TokenCounter, error) {
	switch strings.ToLower(provider) {
	case "openai", "gpt", "gpt-4", "gpt-3.5", "gpt-5":
		return NewOpenAITokenCounter()

	case "anthropic", "claude":
		// For Anthropic, we need to know the specific model
		// Default to claude-3-5-sonnet if not specified
		return NewAnthropicTokenCounter(f.anthropicApiKey, "claude-3-5-sonnet-20241022")

	default:
		return nil, fmt.Errorf("unsupported token counter provider: %s", provider)
	}
}

// CreateTokenCounterForModel creates a token counter based on the model name
func (f *DefaultTokenCounterFactory) CreateTokenCounterForModel(modelName string) (TokenCounter, error) {
	modelName = strings.ToLower(modelName)

	switch {
	// OpenAI models
	case strings.HasPrefix(modelName, "openai/"):
		return NewOpenAITokenCounter()

	// Anthropic models
	case strings.HasPrefix(modelName, "anthropic/"):
		return NewAnthropicTokenCounter(f.anthropicApiKey, modelName)

	// XAI models - use OpenAI-compatible tokenizer for now
	case strings.HasPrefix(modelName, "xai/"):
		return NewOpenAITokenCounter()

	// Default to OpenAI tokenizer for unknown models
	default:
		return NewOpenAITokenCounter()
	}
}

// GetProviderFromModel determines the provider from model name
func GetProviderFromModel(modelName string) string {
	modelName = strings.ToLower(modelName)

	switch {
	case strings.Contains(modelName, "gpt"):
		return "openai"
	case strings.Contains(modelName, "claude"):
		return "anthropic"
	case strings.Contains(modelName, "grok"):
		return "xai"
	default:
		return "openai" // default fallback
	}
}
