package engine

import (
	"context"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/tool"
	goopenai "github.com/openai/openai-go"
	"github.com/pkg/errors"
	"github.com/yukinagae/genkit-go-plugins/plugins/openai"
	"os"
)

type (
	Engine interface {
		NewAgentFromConfig(
			ctx context.Context,
			ac config.AgentConfig,
		) (*entity.Agent, error)
		Run(ctx context.Context, req RunRequest) (*RunResponse, error)
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
	di.Register(Key, func(ctx context.Context, env di.Env) (any, error) {
		conf := di.MustGet[*config.OpenAIConfig](ctx, config.OpenAIConfigKey)
		os.Setenv("OPENAI_API_KEY", conf.OpenAIApiKey)

		if !openai.IsDefinedModel(goopenai.ChatModelGPT4o) {
			if err := openai.Init(ctx, &openai.Config{
				APIKey: conf.OpenAIApiKey,
			}); err != nil {
				return nil, errors.WithStack(err)
			}
		}

		return &engine{
			logger:      di.MustGet[*mylog.Logger](ctx, mylog.Key),
			toolManager: di.MustGet[tool.Manager](ctx, tool.ManagerKey),
		}, nil
	})
}
