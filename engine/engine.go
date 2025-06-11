package engine

import (
	"github.com/firebase/genkit/go/genkit"

	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/tool"
)

type (
	Engine struct {
		logger      *mylog.Logger
		toolManager tool.Manager
		genkit      *genkit.Genkit
	}
)

func NewEngine(
	logger *mylog.Logger,
	toolManager tool.Manager,
	genkit *genkit.Genkit,
) *Engine {
	return &Engine{
		logger:      logger,
		toolManager: toolManager,
		genkit:      genkit,
	}
}
