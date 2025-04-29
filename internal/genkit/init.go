package genkit

import (
	"context"
	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/openai"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/pkg/errors"
	"log/slog"
)

var (
	Key = di.NewKey()
)

func init() {
	di.Register(Key, func(ctx context.Context, c *di.Container) (any, error) {
		conf := di.MustGet[*config.OpenAIConfig](ctx, c, config.OpenAIConfigKey)
		logger := di.MustGet[*slog.Logger](ctx, c, mylog.Key)
		g, err := genkit.Init(
			ctx,
			genkit.WithPlugins(&openai.Plugin{
				APIKey: conf.OpenAIApiKey,
			}),
			genkit.WithDefaultModel("openai/gpt-4o"),
		)

		genkit.RegisterSpanProcessor(g, &loggingSpanProcessor{logger: logger})

		return g, errors.Wrapf(err, "failed to init genkit")
	})
}
