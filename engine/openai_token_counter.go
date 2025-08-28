package engine

import (
	"encoding/json"

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

// CountTokens counts tokens in text using tiktoken
func (o *OpenAITokenCounter) CountTokens(text string) int {
	return len(o.enc.Encode(text, nil, nil))
}

// CountFileTokens counts tokens in file content using OpenAI's vision model algorithm
func (o *OpenAITokenCounter) CountFileTokens(contentType, data string) int {
	return o.imageCalculator.CalculateImageTokens(contentType, data)
}

// CountConversationTokens counts tokens in conversations (text only)
func (o *OpenAITokenCounter) CountConversationTokens(conversations []Conversation) int {
	totalTokens := 0
	for _, conv := range conversations {
		// Serialize conversation to JSON and count tokens
		convJson, _ := json.Marshal(conv)
		totalTokens += o.CountTokens(string(convJson))
	}
	return totalTokens
}

// CountRequestFilesTokens counts tokens in request files
func (o *OpenAITokenCounter) CountRequestFilesTokens(files []File) int {
	totalTokens := 0
	for _, file := range files {
		totalTokens += o.CountFileTokens(file.ContentType, file.Data)
	}
	return totalTokens
}

// ProviderName returns the provider name
func (o *OpenAITokenCounter) ProviderName() string {
	return "openai"
}
