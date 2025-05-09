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
	}

	engine struct {
		logger      *mylog.Logger
		toolManager tool.Manager
		genkit      *genkit.Genkit
	}
)

func init() {
	din.RegisterT(func(c *din.Container) (Engine, error) {
		return &engine{
			logger:      din.MustGet[*mylog.Logger](c, mylog.Key),
			toolManager: din.MustGetT[tool.Manager](c),
			genkit:      din.MustGet[*genkit.Genkit](c, mygenkit.Key),
		}, nil
	})
}
