package genkit

import (
	"context"
	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/openai"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/xai"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/pkg/errors"
	"log/slog"
)

var (
	Key = di.NewKey()
)

func init() {
	di.Register(Key, func(ctx context.Context, c *di.Container) (any, error) {
		var (
			plugins      []genkit.Plugin
			defaultModel string
		)
		{
			conf := di.MustGet[*config.OpenAIConfig](ctx, c, config.OpenAIConfigKey)
			if conf.APIKey != "" {
				plugins = append(plugins, &openai.Plugin{
					APIKey: conf.APIKey,
				})
			}
			defaultModel = "openai/gpt-4o"
		}
		{
			conf := di.MustGet[*config.XAIConfig](ctx, c, config.XAIConfigKey)
			if conf.APIKey != "" {
				plugins = append(plugins, &xai.Plugin{
					APIKey: conf.APIKey,
				})
			}
			defaultModel = "xai/grok-3"
		}
		logConf := di.MustGet[*config.LogConfig](ctx, c, config.LogConfigKey)
		logger := di.MustGet[*slog.Logger](ctx, c, mylog.Key)
		g, err := genkit.Init(
			ctx,
			genkit.WithPlugins(plugins...),
			genkit.WithDefaultModel(defaultModel),
		)

		genkit.RegisterSpanProcessor(g,
			&loggingSpanProcessor{
				verbose: logConf.TraceVerbose,
				logger:  logger,
			},
		)

		return g, errors.Wrapf(err, "failed to init genkit")
	})
}
