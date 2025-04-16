package runner

import (
	"context"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/tool"
	"github.com/pkg/errors"
	"github.com/yukinagae/genkit-go-plugins/plugins/openai"
	"os"
)

type (
	Runner interface {
		NewAgentFromConfig(
			ctx context.Context,
			ac config.AgentConfig,
		) (*entity.Agent, error)
		Run(ctx context.Context, req RunRequest) (*RunResponse, error)
	}

	runner struct {
		logger      *mylog.Logger
		toolManager tool.Manager
	}
)

var (
	_   Runner = (*runner)(nil)
	Key        = di.NewKey()
)

func init() {
	di.Register(Key, func(ctx context.Context, env di.Env) (any, error) {
		conf := di.MustGet[*config.OpenAIConfig](ctx, config.OpenAIConfigKey)
		os.Setenv("OPENAI_API_KEY", conf.OpenAIApiKey)
		if err := openai.Init(ctx, &openai.Config{
			APIKey: conf.OpenAIApiKey,
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		return &runner{
			logger:      di.MustGet[*mylog.Logger](ctx, mylog.Key),
			toolManager: di.MustGet[tool.Manager](ctx, tool.ManagerKey),
		}, nil
	})
}
