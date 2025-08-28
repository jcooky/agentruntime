package engine

// No imports needed for the interface definition

// TokenCounter defines the interface for counting tokens across different model providers
type TokenCounter interface {
	// CountTokens counts tokens in text content
	CountTokens(text string) int

	// CountFileTokens counts tokens in file content (images, PDFs, etc.)
	CountFileTokens(contentType, data string) int

	// CountConversationTokens counts tokens in a list of conversations
	CountConversationTokens(conversations []Conversation) int

	// CountRequestFilesTokens counts tokens in request files
	CountRequestFilesTokens(files []File) int

	// ProviderName returns the name of the token provider
	ProviderName() string
}

// TokenCounterFactory creates TokenCounter for different providers
type TokenCounterFactory interface {
	CreateTokenCounter(provider string) (TokenCounter, error)
}

// SupportedProviders returns list of supported token counting providers
var SupportedProviders = []string{
	"openai",
	"anthropic",
}

// TokenCountRequest represents a request for token counting
type TokenCountRequest struct {
	Text     string         `json:"text,omitempty"`
	Messages []Conversation `json:"messages,omitempty"`
	Files    []File         `json:"files,omitempty"`
	Model    string         `json:"model,omitempty"`
}

// TokenCountResponse represents the response from token counting
type TokenCountResponse struct {
	Tokens   int    `json:"tokens"`
	Provider string `json:"provider"`
	Model    string `json:"model,omitempty"`
}

// EstimateTokens provides a fallback estimation for unsupported providers
func EstimateTokens(text string) int {
	// Rough estimation: ~4 characters per token (English text)
	// This is a fallback for when specific token counters are not available
	return len(text) / 4
}

// EstimateFileTokens provides a fallback estimation for file tokens
func EstimateFileTokens(contentType, data string) int {
	// Base estimation logic for different file types
	switch {
	case contentType == "application/pdf":
		// Estimate ~1 token per 10 bytes for PDF
		return len(data) / 10
	case contentType[:6] == "image/":
		// Estimate ~1 token per 100 bytes for images (rough approximation)
		return len(data) / 100
	case contentType[:5] == "text/":
		// Text files: use standard text estimation
		return EstimateTokens(data)
	default:
		// Generic binary files: conservative estimation
		return len(data) / 50
	}
}
