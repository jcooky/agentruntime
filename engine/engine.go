package engine

import (
	"context"

	"github.com/firebase/genkit/go/genkit"
	mygenkit "github.com/habiliai/agentruntime/internal/genkit"
	"github.com/jcooky/go-din"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/tool"
	"github.com/habiliai/agentruntime/memory"
)

type (
	Engine interface {
		NewAgentFromConfig(
			ctx context.Context,
			ac config.AgentConfig,
		) (*entity.Agent, error)
		Run(ctx context.Context, req RunRequest, output any) (*RunResponse, error)
		Generate(ctx context.Context, req *GenerateRequest, out any, opts ...ai.GenerateOption) (*ai.ModelResponse, error)
		Embed(ctx context.Context, texts ...string) ([][]float32, error)
		IndexKnowledge(ctx context.Context, agentName string, knowledge []map[string]any) error
		RetrieveRelevantKnowledge(ctx context.Context, agentName string, query string, limit int) ([]string, error)
	}

	engine struct {
		logger        *mylog.Logger
		toolManager   tool.Manager
		genkit        *genkit.Genkit
		memoryService memory.Service
	}
)

func init() {
	din.RegisterT(func(c *din.Container) (Engine, error) {
		logger := din.MustGet[*mylog.Logger](c, mylog.Key)

		// Try to get memory service, but don't fail if not available (e.g., in tests)
		var memoryService memory.Service
		if ms, err := din.GetT[memory.Service](c); err == nil {
			memoryService = ms
		} else {
			logger.Warn("memory service not available, RAG functionality will be disabled", "error", err)
		}

		return &engine{
			logger:        logger,
			toolManager:   din.MustGetT[tool.Manager](c),
			genkit:        din.MustGet[*genkit.Genkit](c, mygenkit.Key),
			memoryService: memoryService,
		}, nil
	})
}

// IndexKnowledge indexes knowledge documents for an agent
func (e *engine) IndexKnowledge(ctx context.Context, agentName string, knowledge []map[string]any) error {
	if e.memoryService == nil {
		return nil // Gracefully handle when memory service is not available
	}
	return e.memoryService.IndexKnowledge(ctx, agentName, knowledge)
}

// RetrieveRelevantKnowledge retrieves relevant knowledge chunks based on query
func (e *engine) RetrieveRelevantKnowledge(ctx context.Context, agentName string, query string, limit int) ([]string, error) {
	if e.memoryService == nil {
		return nil, nil // Gracefully handle when memory service is not available
	}
	return e.memoryService.RetrieveRelevantKnowledge(ctx, agentName, query, limit)
}
