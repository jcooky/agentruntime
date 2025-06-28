package engine

import (
	"log/slog"

	"github.com/firebase/genkit/go/genkit"
	tool2 "github.com/habiliai/agentruntime/internal/tool"
	"github.com/habiliai/agentruntime/knowledge"
)

type (
	Engine struct {
		logger           *slog.Logger
		toolManager      tool2.Manager
		genkit           *genkit.Genkit
		knowledgeService knowledge.Service
	}
)

func NewEngine(
	logger *slog.Logger,
	toolManager tool2.Manager,
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
