package engine

import (
	"github.com/firebase/genkit/go/genkit"

	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/tool"
	"github.com/habiliai/agentruntime/memory"
)

type (
	Engine struct {
		logger        *mylog.Logger
		toolManager   tool.Manager
		genkit        *genkit.Genkit
		memoryService memory.Service
	}
)

func NewEngine(
	logger *mylog.Logger,
	toolManager tool.Manager,
	genkit *genkit.Genkit,
	memoryService memory.Service,
) *Engine {
	return &Engine{
		logger:        logger,
		toolManager:   toolManager,
		genkit:        genkit,
		memoryService: memoryService,
	}
}
