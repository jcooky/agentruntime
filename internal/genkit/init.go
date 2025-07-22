package genkit

import (
	"context"
	"log/slog"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/anthropic"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/xai"
	"github.com/jcooky/go-din"
	"github.com/pkg/errors"
)

var (
	Key = din.NewRandomName()
)

func NewGenkit(
	ctx context.Context,
	modelConfig *config.ModelConfig,
	logger *slog.Logger,
	traceVerbose bool,
) (*genkit.Genkit, error) {
	var (
		plugins      []genkit.Plugin
		defaultModel string
	)
	{
		if modelConfig != nil && modelConfig.OpenAIAPIKey != "" {
			plugins = append(plugins, &openai.OpenAI{
				APIKey: modelConfig.OpenAIAPIKey,
			})
			defaultModel = "openai/gpt-4o"
			logger.Info("Loaded OpenAI plugin", "model", defaultModel)
		}
	}
	{
		if modelConfig != nil && modelConfig.XAIAPIKey != "" {
			plugins = append(plugins, &xai.Plugin{
				APIKey: modelConfig.XAIAPIKey,
			})
			defaultModel = "xai/grok-3"
			logger.Info("Loaded XAI plugin", "model", defaultModel)
		}
	}
	{
		if modelConfig != nil && modelConfig.AnthropicAPIKey != "" {
			plugins = append(plugins, &anthropic.Plugin{
				APIKey: modelConfig.AnthropicAPIKey,
			})
			defaultModel = "anthropic/claude-4-sonnet"
			logger.Info("Loaded Anthropic plugin", "model", defaultModel)
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
