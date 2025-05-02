package agentruntime

import (
	"context"
	"io"
	"log/slog"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/tool"
)

type (
	RunRequest      = engine.RunRequest
	RunResponse     = engine.RunResponse
	Conversation    = engine.Conversation
	ToolCall        = engine.ToolCall
	Agent           = entity.Agent
	Participant     = engine.Participant
	MessageExample  = entity.MessageExample
	Tool            = entity.Tool
	ToolConfig      = config.ToolConfig
	GenerateRequest = engine.GenerateRequest

	AIEngine struct {
		engine      engine.Engine
		toolManager tool.Manager

		toolConfig   *ToolConfig
		openAIAPIKey string
		xaiAPIKey    string
		logger       *slog.Logger
	}
)

func (a *AIEngine) Generate(ctx context.Context, req GenerateRequest, out any, opts ...ai.GenerateOption) (*ai.ModelResponse, error) {
	return a.engine.Generate(ctx, &req, out, opts...)
}

func (a *AIEngine) Embed(ctx context.Context, texts ...string) ([][]float32, error) {
	return a.engine.Embed(ctx, texts...)
}

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
			APIKey: e.openAIAPIKey,
		})
	}
	if e.xaiAPIKey != "" {
		di.Set(container, config.XAIConfigKey, &config.XAIConfig{
			APIKey: e.xaiAPIKey,
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

func WithXAIAPIKey(apiKey string) func(e *AIEngine) {
	return func(e *AIEngine) {
		e.xaiAPIKey = apiKey
	}
}

func WithToolConfig(toolConfig *config.ToolConfig) func(e *AIEngine) {
	return func(e *AIEngine) {
		e.toolConfig = toolConfig
	}
}
