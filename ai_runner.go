package agentruntime

import (
	"context"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/runner"
	"log/slog"
)

type (
	RunRequest     = runner.RunRequest
	RunResponse    = runner.RunResponse
	Conversation   = runner.Conversation
	ToolCall       = runner.RunResponseToolcall
	Agent          = entity.Agent
	Participant    = runner.Participant
	MessageExample = entity.MessageExample
	Tool           = entity.Tool

	AIRunner struct {
		runner runner.Runner
	}
	newAIRunnerOption struct {
		openAIAPIKey string
		logger       *slog.Logger
	}
)

func (a *AIRunner) Run(ctx context.Context, req RunRequest) (*RunResponse, error) {
	return a.runner.Run(ctx, req)
}

func NewAIRunner(ctx context.Context, optionFuncs ...func(*newAIRunnerOption)) *AIRunner {
	ctx = di.WithContainer(ctx, di.EnvProd)
	opt := newAIRunnerOption{}
	for _, f := range optionFuncs {
		f(&opt)
	}

	if opt.openAIAPIKey != "" {
		di.Set(ctx, config.OpenAIConfigKey, &config.OpenAIConfig{
			OpenAIApiKey: opt.openAIAPIKey,
		})
	}
	if opt.logger != nil {
		di.Set(ctx, mylog.Key, opt.logger)
	}

	return &AIRunner{
		runner: di.MustGet[runner.Runner](ctx, runner.Key),
	}
}

func WithOpenAIAPIKey(apiKey string) func(*newAIRunnerOption) {
	return func(opt *newAIRunnerOption) {
		opt.openAIAPIKey = apiKey
	}
}

func WithLogger(logger *slog.Logger) func(*newAIRunnerOption) {
	return func(opt *newAIRunnerOption) {
		opt.logger = logger
	}
}
