package agentruntime

import (
	"context"
	"errors"
	"log/slog"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/internal/genkit"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/tool"
)

type (
	AgentRuntime struct {
		engine      engine.Engine
		toolManager tool.Manager
		logger      *slog.Logger

		toolConfig      *config.ToolConfig
		openaiConfig    *config.OpenAIConfig
		xaiConfig       *config.XAIConfig
		anthropicConfig *config.AnthropicConfig
		memoryConfig    *config.MemoryConfig
		logConfig       *config.LogConfig
		traceVerbose    bool
	}
)

func (a *AgentRuntime) Generate(ctx context.Context, req engine.GenerateRequest, out any, opts ...ai.GenerateOption) (*ai.ModelResponse, error) {
	return a.engine.Generate(ctx, &req, out, opts...)
}

func (a *AgentRuntime) Embed(ctx context.Context, texts ...string) ([][]float32, error) {
	return a.engine.Embed(ctx, texts...)
}

func (a *AgentRuntime) Run(ctx context.Context, req engine.RunRequest, out any) (*engine.RunResponse, error) {
	return a.engine.Run(ctx, req, out)
}

func (a *AgentRuntime) Close() {
	a.toolManager.Close()
}

func NewAgentRuntime(ctx context.Context, optionFuncs ...func(*AgentRuntime)) (*AgentRuntime, error) {
	e := &AgentRuntime{
		openaiConfig:    &config.OpenAIConfig{},
		xaiConfig:       &config.XAIConfig{},
		anthropicConfig: &config.AnthropicConfig{},
		memoryConfig:    config.NewMemoryConfig(),
		logConfig:       config.NewLogConfig(),
		toolConfig:      config.NewToolConfig(),
	}
	for _, f := range optionFuncs {
		f(e)
	}

	if e.logger == nil {
		e.logger = mylog.NewLogger(e.logConfig.LogLevel, e.logConfig.LogHandler)
	}

	if e.toolConfig == nil {
		e.toolConfig = config.NewToolConfig()
	}

	if e.openaiConfig == nil && e.xaiConfig == nil && e.anthropicConfig == nil {
		return nil, errors.New("at least one of openai, xai, or anthropic must be configured")
	}

	g, err := genkit.NewGenkit(ctx, e.openaiConfig, e.xaiConfig, e.anthropicConfig, e.logger, e.traceVerbose)
	if err != nil {
		return nil, err
	}
	e.engine = engine.NewEngine(
		e.logger,
		e.toolManager,
		g,
	)

	return e, nil
}

func WithOpenAIAPIKey(apiKey string) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.openaiConfig.APIKey = apiKey
	}
}

func WithLogger(logger *slog.Logger) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.logger = logger
	}
}

func WithTraceVerbose(traceVerbose bool) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.traceVerbose = traceVerbose
	}
}

func WithXAIAPIKey(apiKey string) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.xaiConfig.APIKey = apiKey
	}
}

func WithSerpApiKey(apiKey string) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.toolConfig.SerpApiKey = apiKey
	}
}

func WithOpenWeatherApiKey(apiKey string) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.toolConfig.OpenWeatherApiKey = apiKey
	}
}

func WithLogConfig(logConfig *config.LogConfig) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.logConfig = logConfig
	}
}

func AddMCPServer(serverID string, command string, args []string, env map[string]string) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.toolConfig.MCPServers = append(e.toolConfig.MCPServers, config.MCPServerConfig{
			ID:      serverID,
			Command: command,
			Args:    args,
			Env:     env,
		})
	}
}
