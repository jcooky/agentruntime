package genkit

import (
	"context"
	"log/slog"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/anthropic"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/xai"
	"github.com/jcooky/go-din"
)

var (
	Key = din.NewRandomName()
)

func NewGenkit(
	ctx context.Context,
	openaiConfig *config.OpenAIConfig,
	xaiConfig *config.XAIConfig,
	anthropicConfig *config.AnthropicConfig,
	logger *slog.Logger,
	traceVerbose bool,
) (*genkit.Genkit, error) {
	var (
		plugins      []genkit.Plugin
		defaultModel string
	)
	{
		if openaiConfig != nil && openaiConfig.APIKey != "" {
			plugins = append(plugins, &openai.OpenAI{
				APIKey: openaiConfig.APIKey,
			})
			defaultModel = "openai/gpt-4o"
		}
	}
	{
		if xaiConfig != nil && xaiConfig.APIKey != "" {
			plugins = append(plugins, &xai.Plugin{
				APIKey: xaiConfig.APIKey,
			})
			defaultModel = "xai/grok-3"
		}
	}
	{
		if anthropicConfig != nil && anthropicConfig.APIKey != "" {
			plugins = append(plugins, &anthropic.Plugin{
				APIKey: anthropicConfig.APIKey,
			})
			defaultModel = "anthropic/claude-4-sonnet"
		}
	}
	g, err := genkit.Init(
		ctx,
		genkit.WithPlugins(plugins...),
		genkit.WithDefaultModel(defaultModel),
	)

	genkit.RegisterSpanProcessor(
		g,
		&loggingSpanProcessor{
			verbose: traceVerbose,
			logger:  logger,
		},
	)

	return g, errors.Wrapf(err, "failed to init genkit")
}
