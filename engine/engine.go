package engine

import (
	"log/slog"

	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/tool"
)

type (
	Engine struct {
		logger      *slog.Logger
		toolManager tool.Manager
		genkit      *genkit.Genkit
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
