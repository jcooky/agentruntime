package engine

import (
	"context"
	"encoding/json"

	"github.com/firebase/genkit/go/ai"
	"github.com/pkg/errors"
	"github.com/pkoukk/tiktoken-go"
)

// OpenAITokenCounter implements TokenCounter for OpenAI models using tiktoken
type OpenAITokenCounter struct {
	enc             *tiktoken.Tiktoken
	imageCalculator *ImageTokenCalculator
}

// NewOpenAITokenCounter creates a new OpenAI token counter
func NewOpenAITokenCounter() (*OpenAITokenCounter, error) {
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tiktoken encoding")
	}

	return &OpenAITokenCounter{
		enc:             enc,
		imageCalculator: NewImageTokenCalculator(),
	}, nil
}

// CountConversationTokens counts tokens in conversations using tiktoken
func (o *OpenAITokenCounter) CountConversationTokens(ctx context.Context, history []*ai.Message) int {
	totalTokens := 0
	for _, msg := range history {
		// Count tokens for each part of the message
		for _, part := range msg.Content {
			if part.IsText() {
				totalTokens += len(o.enc.Encode(part.Text, nil, nil))
			} else if part.IsToolRequest() {
				// Count tokens for tool requests
				toolReqJson, _ := json.Marshal(part.ToolRequest)
				totalTokens += len(o.enc.Encode(string(toolReqJson), nil, nil))
			} else if part.IsToolResponse() {
				// Count tokens for tool responses
				toolRespJson, _ := json.Marshal(part.ToolResponse)
				totalTokens += len(o.enc.Encode(string(toolRespJson), nil, nil))
			} else if part.IsMedia() {
				// For media content, use a simplified estimation
				// In a real implementation, you might want to use vision model token calculation
				totalTokens += EstimateFileTokens(part.ContentType, part.Text)
			}
		}
	}
	return totalTokens
}

// ProviderName returns the provider name
func (o *OpenAITokenCounter) ProviderName() string {
	return "openai"
}
