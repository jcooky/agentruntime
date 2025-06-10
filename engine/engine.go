package engine

import (
	"context"

	"github.com/firebase/genkit/go/genkit"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/tool"
)

type (
	Engine interface {
		Run(ctx context.Context, req RunRequest, output any) (*RunResponse, error)
		Generate(ctx context.Context, req *GenerateRequest, out any, opts ...ai.GenerateOption) (*ai.ModelResponse, error)
		Embed(ctx context.Context, texts ...string) ([][]float32, error)
	}

	engine struct {
		logger      *mylog.Logger
		toolManager tool.Manager
		genkit      *genkit.Genkit
	}
)

func NewEngine(
	logger *mylog.Logger,
	toolManager tool.Manager,
	genkit *genkit.Genkit,
) Engine {
	return &engine{
		logger:      logger,
		toolManager: toolManager,
		genkit:      genkit,
	}
}
