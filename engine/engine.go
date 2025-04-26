package engine

import (
	"context"
	"os"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/tool"
	goopenai "github.com/openai/openai-go"
	"github.com/pkg/errors"
	"github.com/yukinagae/genkit-go-plugins/plugins/openai"
)

type (
	Engine interface {
		NewAgentFromConfig(
			ctx context.Context,
			ac config.AgentConfig,
		) (*entity.Agent, error)
		Run(ctx context.Context, req RunRequest, output any) (*RunResponse, error)
		Generate(ctx context.Context, req GenerateRequest, output any, opts ...ai.GenerateOption) (*ai.GenerateResponse, error)
		Embed(ctx context.Context, texts ...string) ([][]float32, error)
	}

	engine struct {
		logger      *mylog.Logger
		toolManager tool.Manager
	}
)

var (
	_   Engine = (*engine)(nil)
	Key        = di.NewKey()
)

func init() {
	di.Register(Key, func(ctx context.Context, c *di.Container) (any, error) {
		conf := di.MustGet[*config.OpenAIConfig](ctx, c, config.OpenAIConfigKey)
		os.Setenv("OPENAI_API_KEY", conf.OpenAIApiKey)

		if !openai.IsDefinedModel(goopenai.ChatModelGPT4o) {
			if err := openai.Init(ctx, &openai.Config{
				APIKey: conf.OpenAIApiKey,
			}); err != nil {
				return nil, errors.WithStack(err)
			}
		}

		return &engine{
			logger:      di.MustGet[*mylog.Logger](ctx, c, mylog.Key),
			toolManager: di.MustGet[tool.Manager](ctx, c, tool.ManagerKey),
		}, nil
	})
}
