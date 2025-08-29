package engine

import (
	"log/slog"

	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/tool"
)

type (
	Engine struct {
		logger                 *slog.Logger
		toolManager            tool.Manager
		genkit                 *genkit.Genkit
		conversationSummarizer *ConversationSummarizer
	}
)

func NewEngine(
	logger *slog.Logger,
	toolManager tool.Manager,
	genkit *genkit.Genkit,
) *Engine {
	return &Engine{
		logger:      logger,
		toolManager: toolManager,
		genkit:      genkit,
	}
}

func NewEngineWithSummarizer(
	logger *slog.Logger,
	toolManager tool.Manager,
	genkit *genkit.Genkit,
	modelConfig *config.ModelConfig,
) (*Engine, error) {
	summarizer := NewConversationSummarizer(genkit, &modelConfig.ConversationSummary)

	return &Engine{
		logger:                 logger,
		toolManager:            toolManager,
		genkit:                 genkit,
		conversationSummarizer: summarizer,
	}, nil
}
