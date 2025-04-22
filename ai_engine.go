package agentruntime

import (
	"context"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"log/slog"
)

type (
	RunRequest     = engine.RunRequest
	RunResponse    = engine.RunResponse
	Conversation   = engine.Conversation
	ToolCall       = engine.RunResponseToolcall
	Agent          = entity.Agent
	Participant    = engine.Participant
	MessageExample = entity.MessageExample
	Tool           = entity.Tool

	AIEngine struct {
		engine engine.Engine

		openAIAPIKey string
		logger       *slog.Logger
	}
)

func (a *AIEngine) Run(ctx context.Context, req RunRequest) (*RunResponse, error) {
	return a.engine.Run(ctx, req)
}

func NewAIEngine(ctx context.Context, optionFuncs ...func(*AIEngine)) *AIEngine {
	ctx = di.WithContainer(ctx, di.EnvProd)
	e := &AIEngine{}
	for _, f := range optionFuncs {
		f(e)
	}
	if e.openAIAPIKey != "" {
		di.Set(ctx, config.OpenAIConfigKey, &config.OpenAIConfig{
			OpenAIApiKey: e.openAIAPIKey,
		})
	}
	if e.logger != nil {
		di.Set(ctx, mylog.Key, e.logger)
	}

	e.engine = di.MustGet[engine.Engine](ctx, engine.Key)
	return e
}

func WithOpenAIAPIKey(apiKey string) func(e *AIEngine) {
	return func(e *AIEngine) {
		e.openAIAPIKey = apiKey
	}
}

func WithLogger(logger *slog.Logger) func(e *AIEngine) {
	return func(e *AIEngine) {
		e.logger = logger
	}
}
