package engine

import (
	"log/slog"

	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/knowledge"
	"github.com/habiliai/agentruntime/tool"
)

type (
	Engine struct {
		logger           *slog.Logger
		toolManager      tool.Manager
		genkit           *genkit.Genkit
		knowledgeService knowledge.Service
	}
)

func NewEngine(
	logger *slog.Logger,
	toolManager tool.Manager,
	genkit *genkit.Genkit,
	knowledgeService knowledge.Service,
) *Engine {
	return &Engine{
		logger:           logger,
		toolManager:      toolManager,
		genkit:           genkit,
		knowledgeService: knowledgeService,
	}
}
