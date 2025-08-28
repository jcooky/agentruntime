package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// AnthropicTokenCounter implements TokenCounter for Anthropic models using the count_tokens API
type AnthropicTokenCounter struct {
	client  *http.Client
	apiKey  string
	baseURL string
	model   string
}

// NewAnthropicTokenCounter creates a new Anthropic token counter
func NewAnthropicTokenCounter(anthropicApiKey string, model string) (*AnthropicTokenCounter, error) {
	return &AnthropicTokenCounter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:  anthropicApiKey,
		baseURL: "https://api.anthropic.com/v1",
		model:   model,
	}, nil
}

// anthropicTokenRequest represents the request to Anthropic's count_tokens API
type anthropicTokenRequest struct {
	Model    string                  `json:"model"`
	Messages []anthropicMessageParam `json:"messages,omitempty"`
	System   string                  `json:"system,omitempty"`
}

type anthropicMessageParam struct {
	Role    string                  `json:"role"`
	Content []anthropicContentParam `json:"content"`
}

type anthropicContentParam struct {
	Type   string                `json:"type"`
	Text   string                `json:"text,omitempty"`
	Source *anthropicSourceParam `json:"source,omitempty"`
}

type anthropicSourceParam struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// anthropicTokenResponse represents the response from Anthropic's count_tokens API
type anthropicTokenResponse struct {
	InputTokens int `json:"input_tokens"`
}

// CountTokens counts tokens in text using Anthropic's API
func (a *AnthropicTokenCounter) CountTokens(text string) int {
	if text == "" {
		return 0
	}

	// Create a simple message for token counting
	req := anthropicTokenRequest{
		Model: a.model,
		Messages: []anthropicMessageParam{
			{
				Role: "user",
				Content: []anthropicContentParam{
					{
						Type: "text",
						Text: text,
					},
				},
			},
		},
	}

	tokens, err := a.callCountTokensAPI(context.Background(), req)
	if err != nil {
		// Fallback to estimation if API fails
		return EstimateTokens(text)
	}

	return tokens
}

// CountFileTokens counts tokens in file content using Anthropic's API
func (a *AnthropicTokenCounter) CountFileTokens(contentType, data string) int {
	if data == "" {
		return 0
	}

	// Only handle images through API, fallback for other file types
	if contentType[:6] != "image/" {
		return EstimateFileTokens(contentType, data)
	}

	req := anthropicTokenRequest{
		Model: a.model,
		Messages: []anthropicMessageParam{
			{
				Role: "user",
				Content: []anthropicContentParam{
					{
						Type: "image",
						Source: &anthropicSourceParam{
							Type:      "base64",
							MediaType: contentType,
							Data:      data,
						},
					},
				},
			},
		},
	}

	tokens, err := a.callCountTokensAPI(context.Background(), req)
	if err != nil {
		// Fallback to estimation if API fails
		return EstimateFileTokens(contentType, data)
	}

	return tokens
}

// CountConversationTokens counts tokens in conversations using Anthropic's API
func (a *AnthropicTokenCounter) CountConversationTokens(conversations []Conversation) int {
	if len(conversations) == 0 {
		return 0
	}

	// Convert conversations to Anthropic message format
	messages := make([]anthropicMessageParam, 0, len(conversations))

	for _, conv := range conversations {
		if conv.Text == "" {
			continue
		}

		// Map user roles
		role := "user"
		if conv.User == "assistant" || conv.User == "bot" {
			role = "assistant"
		}

		messages = append(messages, anthropicMessageParam{
			Role: role,
			Content: []anthropicContentParam{
				{
					Type: "text",
					Text: conv.Text,
				},
			},
		})
	}

	if len(messages) == 0 {
		return 0
	}

	req := anthropicTokenRequest{
		Model:    a.model,
		Messages: messages,
	}

	tokens, err := a.callCountTokensAPI(context.Background(), req)
	if err != nil {
		// Fallback to estimation if API fails
		totalText := ""
		for _, conv := range conversations {
			totalText += conv.Text + " "
		}
		return EstimateTokens(totalText)
	}

	return tokens
}

// CountRequestFilesTokens counts tokens in request files
func (a *AnthropicTokenCounter) CountRequestFilesTokens(files []File) int {
	totalTokens := 0
	for _, file := range files {
		totalTokens += a.CountFileTokens(file.ContentType, file.Data)
	}
	return totalTokens
}

// callCountTokensAPI calls Anthropic's count_tokens API
func (a *AnthropicTokenCounter) callCountTokensAPI(ctx context.Context, req anthropicTokenRequest) (int, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return 0, errors.Wrap(err, "failed to marshal request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/messages/count_tokens", bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, errors.Wrap(err, "failed to create request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return 0, errors.Wrap(err, "failed to make request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, errors.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp anthropicTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return 0, errors.Wrap(err, "failed to decode response")
	}

	return tokenResp.InputTokens, nil
}

// ProviderName returns the provider name
func (a *AnthropicTokenCounter) ProviderName() string {
	return "anthropic"
}

// SetModel updates the model for token counting
func (a *AnthropicTokenCounter) SetModel(model string) {
	a.model = model
}
