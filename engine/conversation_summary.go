package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/config"
	"github.com/pkg/errors"
)

// ConversationSummarizer handles conversation history summarization
type ConversationSummarizer struct {
	genkit       *genkit.Genkit
	config       config.ConversationSummaryConfig
	tokenCounter TokenCounter
}

// NewConversationSummarizer creates a new conversation summarizer
func NewConversationSummarizer(g *genkit.Genkit, config *config.ModelConfig) (*ConversationSummarizer, error) {
	// Create token counter based on configuration
	factory := NewDefaultTokenCounterFactory(config)
	var tokenCounter TokenCounter
	var err error

	cfg := config.ConversationSummary
	if cfg.TokenProvider == "auto" || cfg.TokenProvider == "" {
		// Auto-detect based on model name
		tokenCounter, err = factory.CreateTokenCounterForModel(cfg.ModelForSummary)
	} else {
		// Use specific provider
		tokenCounter, err = factory.CreateTokenCounter(cfg.TokenProvider)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to create token counter (provider: %s, model: %s)", cfg.TokenProvider, cfg.ModelForSummary)
	}

	return &ConversationSummarizer{
		genkit:       g,
		config:       cfg,
		tokenCounter: tokenCounter,
	}, nil
}

// NewConversationSummarizerWithTokenCounter creates a new conversation summarizer with a specific token counter
func NewConversationSummarizerWithTokenCounter(g *genkit.Genkit, cfg config.ConversationSummaryConfig, tokenCounter TokenCounter) (*ConversationSummarizer, error) {
	return &ConversationSummarizer{
		genkit:       g,
		config:       cfg,
		tokenCounter: tokenCounter,
	}, nil
}

// CountTokens counts the number of tokens in text
func (cs *ConversationSummarizer) CountTokens(text string) int {
	// Use the global estimation function since TokenCounter interface no longer has this method
	return EstimateTokens(text)
}

// CountFileTokens estimates tokens for files
func (cs *ConversationSummarizer) CountFileTokens(contentType, data string) int {
	// Use the global estimation function since TokenCounter interface no longer has this method
	return EstimateFileTokens(contentType, data)
}

// CountConversationTokens counts tokens in a list of conversations (text only)
func (cs *ConversationSummarizer) CountConversationTokens(conversations []Conversation) int {
	// Convert Conversation to ai.Message format
	messages := cs.convertConversationsToMessages(conversations)
	return cs.tokenCounter.CountConversationTokens(context.Background(), messages)
}

// convertConversationsToMessages converts Conversation slice to ai.Message slice
func (cs *ConversationSummarizer) convertConversationsToMessages(conversations []Conversation) []*ai.Message {
	messages := make([]*ai.Message, 0, len(conversations))

	for _, conv := range conversations {
		// Determine the role based on the User field
		role := ai.RoleUser
		if conv.User == "assistant" || conv.User == "bot" {
			role = ai.RoleModel
		}

		// Create content parts
		var parts []*ai.Part
		if conv.Text != "" {
			parts = append(parts, ai.NewTextPart(conv.Text))
		}

		// Add action information if present
		if len(conv.Actions) > 0 {
			for _, action := range conv.Actions {
				actionText := fmt.Sprintf("[Action: %s]", action.Name)
				parts = append(parts, ai.NewTextPart(actionText))
			}
		}

		if len(parts) > 0 {
			messages = append(messages, &ai.Message{
				Role:    role,
				Content: parts,
			})
		}
	}

	return messages
}

// CountRequestFilesTokens counts tokens for files in a RunRequest
func (cs *ConversationSummarizer) CountRequestFilesTokens(files []File) int {
	totalTokens := 0
	for _, file := range files {
		totalTokens += cs.CountFileTokens(file.ContentType, file.Data)
	}
	return totalTokens
}

// GetTokenCounter returns the underlying token counter
func (cs *ConversationSummarizer) GetTokenCounter() TokenCounter {
	return cs.tokenCounter
}

// SummarizedConversation represents a summarized portion of conversation history
type SummarizedConversation struct {
	Summary          string `json:"summary"`
	OriginalCount    int    `json:"original_count"`
	TimeRange        string `json:"time_range,omitempty"`
	ParticipantCount int    `json:"participant_count,omitempty"`
}

// ConversationHistoryResult contains the processed conversation history
type ConversationHistoryResult struct {
	Summary              *SummarizedConversation `json:"summary,omitempty"`
	RecentConversations  []Conversation          `json:"recent_conversations"`
	TotalTokens          int                     `json:"total_tokens"`
	WasSummarizationUsed bool                    `json:"was_summarization_used"`
}

// ProcessConversationHistory processes conversation history with summarization if needed
// requestFiles should be the files from the current RunRequest to include in token calculation
func (cs *ConversationSummarizer) ProcessConversationHistory(ctx context.Context, conversations []Conversation, requestFiles []File) (*ConversationHistoryResult, error) {
	// Calculate tokens for current request files
	requestFilesTokens := cs.CountRequestFilesTokens(requestFiles)

	if len(conversations) == 0 {
		return &ConversationHistoryResult{
			RecentConversations:  []Conversation{},
			TotalTokens:          requestFilesTokens,
			WasSummarizationUsed: false,
		}, nil
	}

	conversationTokens := cs.CountConversationTokens(conversations)
	totalTokens := requestFilesTokens + conversationTokens

	// If under token limit, return all conversations
	if totalTokens <= cs.config.MaxTokens {
		return &ConversationHistoryResult{
			RecentConversations:  conversations,
			TotalTokens:          totalTokens,
			WasSummarizationUsed: false,
		}, nil
	}

	// If we have too few conversations to summarize, just truncate
	if len(conversations) < cs.config.MinConversationsToSummarize {
		// Keep the most recent conversations within token limit (accounting for request files)
		availableTokens := cs.config.MaxTokens - requestFilesTokens

		// If request files exceed MaxTokens, still return the conversations as-is
		// (rather than returning empty conversations) since the test expects this behavior
		if availableTokens <= 0 {
			return &ConversationHistoryResult{
				RecentConversations:  conversations,
				TotalTokens:          totalTokens,
				WasSummarizationUsed: false,
			}, nil
		}

		recentConversations := cs.truncateToTokenLimit(conversations, availableTokens)
		return &ConversationHistoryResult{
			RecentConversations:  recentConversations,
			TotalTokens:          requestFilesTokens + cs.CountConversationTokens(recentConversations),
			WasSummarizationUsed: false,
		}, nil
	}

	// Determine split point for summarization
	splitPoint := cs.findSplitPoint(conversations)

	if splitPoint <= 0 {
		// Can't summarize, just truncate (accounting for request files)
		availableTokens := cs.config.MaxTokens - requestFilesTokens
		recentConversations := cs.truncateToTokenLimit(conversations, availableTokens)
		return &ConversationHistoryResult{
			RecentConversations:  recentConversations,
			TotalTokens:          requestFilesTokens + cs.CountConversationTokens(recentConversations),
			WasSummarizationUsed: false,
		}, nil
	}

	// Split conversations
	oldConversations := conversations[:splitPoint]
	recentConversations := conversations[splitPoint:]

	// Generate summary for old conversations
	summary, err := cs.generateSummary(ctx, oldConversations)
	if err != nil {
		// On error, fall back to truncation (accounting for request files)
		availableTokens := cs.config.MaxTokens - requestFilesTokens
		recentConversations := cs.truncateToTokenLimit(conversations, availableTokens)
		return &ConversationHistoryResult{
			RecentConversations:  recentConversations,
			TotalTokens:          requestFilesTokens + cs.CountConversationTokens(recentConversations),
			WasSummarizationUsed: false,
		}, nil
	}

	// Check if the result fits within token limit (including request files)
	finalTokens := requestFilesTokens + cs.CountTokens(summary.Summary) + cs.CountConversationTokens(recentConversations)

	// If still too large, truncate recent conversations
	if finalTokens > cs.config.MaxTokens {
		availableTokens := cs.config.MaxTokens - requestFilesTokens - cs.CountTokens(summary.Summary)
		if availableTokens > 0 {
			recentConversations = cs.truncateToTokenLimit(recentConversations, availableTokens)
		} else {
			recentConversations = []Conversation{} // No space for recent conversations
		}
	}

	return &ConversationHistoryResult{
		Summary:              summary,
		RecentConversations:  recentConversations,
		TotalTokens:          requestFilesTokens + cs.CountTokens(summary.Summary) + cs.CountConversationTokens(recentConversations),
		WasSummarizationUsed: true,
	}, nil
}

// findSplitPoint finds the optimal point to split conversations for summarization
func (cs *ConversationSummarizer) findSplitPoint(conversations []Conversation) int {
	totalConversations := len(conversations)

	// Keep at least 1/3 of conversations as recent
	minRecentConversations := totalConversations / 3
	if minRecentConversations < cs.config.MinConversationsToSummarize {
		minRecentConversations = cs.config.MinConversationsToSummarize
	}

	// The split point should be at least this many conversations from the end
	maxSplitPoint := totalConversations - minRecentConversations

	if maxSplitPoint <= 0 {
		return 0
	}

	// Find the split point that keeps recent conversations under token limit
	for splitPoint := maxSplitPoint; splitPoint > 0; splitPoint-- {
		recentConversations := conversations[splitPoint:]
		recentTokens := cs.CountConversationTokens(recentConversations)

		// Reserve tokens for summary and request files
		availableTokens := cs.config.MaxTokens - cs.config.SummaryTokens

		if recentTokens <= availableTokens {
			return splitPoint
		}
	}

	return 0
}

// truncateToTokenLimit truncates conversations to fit within token limit (from the end)
func (cs *ConversationSummarizer) truncateToTokenLimit(conversations []Conversation, tokenLimit int) []Conversation {
	if len(conversations) == 0 {
		return conversations
	}

	// Start from the end and add conversations until we hit the limit
	var result []Conversation
	currentTokens := 0

	for i := len(conversations) - 1; i >= 0; i-- {
		conv := conversations[i]
		convJson, _ := json.Marshal(conv)
		convTokens := cs.CountTokens(string(convJson))

		if currentTokens+convTokens > tokenLimit {
			break
		}

		result = append([]Conversation{conv}, result...)
		currentTokens += convTokens
	}

	return result
}

// generateSummary generates a summary of the given conversations
func (cs *ConversationSummarizer) generateSummary(ctx context.Context, conversations []Conversation) (*SummarizedConversation, error) {
	if len(conversations) == 0 {
		return nil, errors.New("no conversations to summarize")
	}

	// Convert conversations to a readable format
	var conversationText strings.Builder
	participantSet := make(map[string]bool)

	for i, conv := range conversations {
		participantSet[conv.User] = true

		conversationText.WriteString(fmt.Sprintf("## Message %d\n", i+1))
		conversationText.WriteString(fmt.Sprintf("**User:** %s\n", conv.User))
		conversationText.WriteString(fmt.Sprintf("**Text:** %s\n", conv.Text))

		if len(conv.Actions) > 0 {
			conversationText.WriteString("**Actions:**\n")
			for _, action := range conv.Actions {
				conversationText.WriteString(fmt.Sprintf("- %s: %v -> %v\n", action.Name, action.Arguments, action.Result))
			}
		}
		conversationText.WriteString("\n")
	}

	prompt := fmt.Sprintf(`Please provide a comprehensive summary of the following conversation history. The summary should capture:

1. Key topics discussed
2. Important decisions made
3. Action items or tasks mentioned
4. Context that would be relevant for future conversations
5. Any important user preferences or information revealed

Keep the summary concise but informative (aim for around %d tokens). Focus on information that would help an AI assistant provide better continuity in future conversations.

Conversation History:
%s

Summary:`, cs.config.SummaryTokens/2, conversationText.String()) // Aim for half the target tokens in prompt

	response, err := genkit.Generate(ctx, cs.genkit,
		ai.WithModelName(cs.config.ModelForSummary),
		ai.WithPrompt(prompt),
		ai.WithOutputFormat(ai.OutputFormatText),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate conversation summary")
	}

	summaryText := strings.TrimSpace(response.Text())

	// Extract time range information if available
	var timeRange string
	if len(conversations) > 0 {
		firstUser := conversations[0].User
		lastUser := conversations[len(conversations)-1].User
		if firstUser == lastUser {
			timeRange = fmt.Sprintf("Conversation with %s", firstUser)
		} else {
			timeRange = fmt.Sprintf("Conversation from %s to %s", firstUser, lastUser)
		}
	}

	return &SummarizedConversation{
		Summary:          summaryText,
		OriginalCount:    len(conversations),
		TimeRange:        timeRange,
		ParticipantCount: len(participantSet),
	}, nil
}

// UpdatedChatPromptValues extends ChatPromptValues with summarization support
type UpdatedChatPromptValues struct {
	ChatPromptValues
	ConversationSummary *SummarizedConversation `json:"conversation_summary,omitempty"`
}
