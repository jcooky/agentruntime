package agentruntime

import (
	"context"
	"io"
	"log/slog"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/tool"
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
	ToolConfig     = config.ToolConfig

	AIEngine struct {
		engine      engine.Engine
		toolManager tool.Manager

		toolConfig   *ToolConfig
		openAIAPIKey string
		logger       *slog.Logger
	}
)

func (a *AIEngine) Run(ctx context.Context, req RunRequest, out any) (*RunResponse, error) {
	return a.engine.Run(ctx, req, out)
}

func (a *AIEngine) CreateAgentFromYaml(ctx context.Context, yamlFile io.Reader) (*Agent, error) {
	yamlCfg, err := config.LoadAgentFromFile(yamlFile)
	if err != nil {
		return nil, err
	}

	for name, mcpServer := range yamlCfg.MCPServers {
		if err := a.toolManager.RegisterMCPTool(ctx, tool.RegisterMCPToolRequest{
			ServerName: name,
			Command:    mcpServer.Command,
			Args:       mcpServer.Args,
			Env:        mcpServer.Env,
		}); err != nil {
			return nil, err
		}
	}
	return a.engine.NewAgentFromConfig(ctx, yamlCfg)
}

func NewAIEngine(ctx context.Context, optionFuncs ...func(*AIEngine)) *AIEngine {
	container := di.NewContainer(di.EnvProd)
	e := &AIEngine{}
	for _, f := range optionFuncs {
		f(e)
	}
	if e.openAIAPIKey != "" {
		di.Set(container, config.OpenAIConfigKey, &config.OpenAIConfig{
			OpenAIApiKey: e.openAIAPIKey,
		})
	}
	if e.logger != nil {
		di.Set(container, mylog.Key, e.logger)
	}
	if e.toolConfig != nil {
		di.Set(container, config.ToolConfigKey, e.toolConfig)
	}

	e.engine = di.MustGet[engine.Engine](ctx, container, engine.Key)
	e.toolManager = di.MustGet[tool.Manager](ctx, container, tool.ManagerKey)
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

func WithToolConfig(toolConfig *config.ToolConfig) func(e *AIEngine) {
	return func(e *AIEngine) {
		e.toolConfig = toolConfig
	}
}
