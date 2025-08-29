package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
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

// CountConversationTokens counts tokens in conversations using Anthropic's API
func (a *AnthropicTokenCounter) CountConversationTokens(ctx context.Context, history []*ai.Message) int {
	if len(history) == 0 {
		return 0
	}

	// Convert ai.Message to Anthropic message format
	messages, systemPrompt := a.convertMessages(history)

	if len(messages) == 0 {
		return 0
	}

	req := anthropicTokenRequest{
		Model:    a.model,
		Messages: messages,
		System:   systemPrompt,
	}

	tokens, err := a.callCountTokensAPI(ctx, req)
	if err != nil {
		// Fallback to estimation if API fails
		totalText := ""
		for _, msg := range history {
			for _, part := range msg.Content {
				if part.IsText() {
					totalText += part.Text + " "
				}
			}
		}
		return EstimateTokens(totalText)
	}

	return tokens
}

// convertMessages converts ai.Message to anthropicMessageParam format
func (a *AnthropicTokenCounter) convertMessages(messages []*ai.Message) ([]anthropicMessageParam, string) {
	var anthropicMessages []anthropicMessageParam
	var systemPrompts []string

	for _, msg := range messages {
		switch msg.Role {
		case ai.RoleSystem:
			// Extract system prompts
			for _, part := range msg.Content {
				if part.IsText() && part.Text != "" {
					systemPrompts = append(systemPrompts, part.Text)
				}
			}
		case ai.RoleUser:
			content := a.convertContent(msg.Content)
			if len(content) > 0 {
				anthropicMessages = append(anthropicMessages, anthropicMessageParam{
					Role:    "user",
					Content: content,
				})
			}
		case ai.RoleModel:
			content := a.convertContent(msg.Content)
			if len(content) > 0 {
				anthropicMessages = append(anthropicMessages, anthropicMessageParam{
					Role:    "assistant",
					Content: content,
				})
			}
		case ai.RoleTool:
			// Tool messages are treated as user messages in Anthropic API
			content := a.convertContent(msg.Content)
			if len(content) > 0 {
				anthropicMessages = append(anthropicMessages, anthropicMessageParam{
					Role:    "user",
					Content: content,
				})
			}
		}
	}

	// Join system prompts
	systemPrompt := ""
	if len(systemPrompts) > 0 {
		systemPrompt = strings.Join(systemPrompts, "\n")
	}

	return anthropicMessages, systemPrompt
}

// convertContent converts ai.Part to anthropicContentParam
func (a *AnthropicTokenCounter) convertContent(parts []*ai.Part) []anthropicContentParam {
	var content []anthropicContentParam

	for _, part := range parts {
		if part.IsText() {
			content = append(content, anthropicContentParam{
				Type: "text",
				Text: part.Text,
			})
		} else if part.IsMedia() {
			// For token counting, we'll convert media to text representation
			// This is a simplified approach - in practice you might want to handle media differently
			content = append(content, anthropicContentParam{
				Type: "text",
				Text: "[Media content]",
			})
		} else if part.IsToolRequest() {
			// Convert tool requests to text for token counting
			content = append(content, anthropicContentParam{
				Type: "text",
				Text: "[Tool request: " + part.ToolRequest.Name + "]",
			})
		} else if part.IsToolResponse() {
			// Convert tool responses to text for token counting
			content = append(content, anthropicContentParam{
				Type: "text",
				Text: "[Tool response]",
			})
		}
	}

	return content
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
