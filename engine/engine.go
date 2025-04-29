package engine

import (
	"context"
	"github.com/firebase/genkit/go/genkit"
	mygenkit "github.com/habiliai/agentruntime/internal/genkit"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/tool"
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

var (
	_   Engine = (*engine)(nil)
	Key        = di.NewKey()
)

func init() {
	di.Register(Key, func(ctx context.Context, c *di.Container) (any, error) {
		return &engine{
			logger:      di.MustGet[*mylog.Logger](ctx, c, mylog.Key),
			toolManager: di.MustGet[tool.Manager](ctx, c, tool.ManagerKey),
			genkit:      di.MustGet[*genkit.Genkit](ctx, c, mygenkit.Key),
		}, nil
	})
}
