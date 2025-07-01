package agentruntime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/genkit"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/tool"
	"github.com/habiliai/agentruntime/knowledge"
)

type (
	AgentRuntime struct {
		engine           *engine.Engine
		toolManager      tool.Manager
		logger           *slog.Logger
		agent            *entity.Agent
		knowledgeService knowledge.Service

		modelConfig     *config.ModelConfig
		knowledgeConfig *config.KnowledgeConfig
		logConfig       *config.LogConfig
	}
	Option func(*AgentRuntime)
)

func (a *AgentRuntime) Agent() *entity.Agent {
	return a.agent
}

func (a *AgentRuntime) Generate(ctx context.Context, req engine.GenerateRequest, out any, opts ...ai.GenerateOption) (*ai.ModelResponse, error) {
	return a.engine.Generate(ctx, &req, out, opts...)
}

func (a *AgentRuntime) Embed(ctx context.Context, texts ...string) ([][]float32, error) {
	return a.engine.Embed(ctx, texts...)
}

func (a *AgentRuntime) Run(ctx context.Context, req engine.RunRequest, out any) (*engine.RunResponse, error) {
	return a.engine.Run(ctx, *a.agent, req, out)
}

func (a *AgentRuntime) Close() {
	a.toolManager.Close()
}

func NewAgentRuntime(ctx context.Context, optionFuncs ...Option) (*AgentRuntime, error) {
	e := &AgentRuntime{
		modelConfig:     &config.ModelConfig{},
		knowledgeConfig: config.NewKnowledgeConfig(),
		logConfig:       config.NewLogConfig(),
	}
	for _, f := range optionFuncs {
		f(e)
	}

	if e.logger == nil {
		e.logger = mylog.NewLogger(e.logConfig.LogLevel, e.logConfig.LogHandler)
	}

	if e.agent == nil {
		return nil, errors.New("agent is required")
	}

	if e.modelConfig == nil {
		return nil, errors.New("model config is required")
	}

	g, err := genkit.NewGenkit(ctx, e.modelConfig, e.logger, e.modelConfig.TraceVerbose)
	if err != nil {
		return nil, err
	}

	if e.knowledgeService == nil {
		e.knowledgeService, err = knowledge.NewServiceWithStore(ctx, e.knowledgeConfig, e.modelConfig, e.logger, knowledge.NewInMemoryStore())
		if err != nil {
			return nil, err
		}
	}

	e.toolManager, err = tool.NewToolManager(ctx, e.agent.Skills, e.logger, g, e.knowledgeService)
	if err != nil {
		return nil, err
	}

	if len(e.agent.Knowledge) > 0 {
		// Index knowledge for RAG if available
		knowledgeId := fmt.Sprintf("%s-knowledge", e.agent.Name)
		if _, err := e.knowledgeService.IndexKnowledgeFromMap(ctx, knowledgeId, e.agent.Knowledge); err != nil {
			e.logger.Warn("failed to index knowledge for agent - agent will work without RAG functionality",
				"agent", e.agent.Name,
				"error", err)
			// Continue without failing agent creation
		}
	}

	e.engine = engine.NewEngine(
		e.logger,
		e.toolManager,
		g,
		e.knowledgeService,
	)

	return e, nil
}

func WithOpenAIAPIKey(apiKey string) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.modelConfig.OpenAIAPIKey = apiKey
	}
}

func WithLogger(logger *slog.Logger) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.logger = logger
	}
}

func WithTraceVerbose(traceVerbose bool) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.modelConfig.TraceVerbose = traceVerbose
	}
}

func WithXAIAPIKey(apiKey string) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.modelConfig.XAIAPIKey = apiKey
	}
}

func WithAnthropicAPIKey(apiKey string) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.modelConfig.AnthropicAPIKey = apiKey
	}
}

func WithLogConfig(logConfig *config.LogConfig) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.logConfig = logConfig
	}
}

func WithAgent(agent entity.Agent) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.agent = &agent
	}
}

func WithKnowledgeService(knowledgeService knowledge.Service) func(e *AgentRuntime) {
	return func(e *AgentRuntime) {
		e.knowledgeService = knowledgeService
	}
}
