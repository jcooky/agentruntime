package tool

import (
	"context"
	"sync"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/mylog"
	mcpclient "github.com/mark3labs/mcp-go/client"
)

type (
	LocalToolService interface {
		GetWeather(ctx context.Context, req *GetWeatherRequest) (*GetWeatherResponse, error)
		DoneAgent(_ context.Context, req *DoneAgentRequest) (*DoneAgentResponse, error)
		Search(ctx context.Context, req *WebSearchRequest) ([]any, error)
	}

	manager struct {
		logger *mylog.Logger
		config *config.ToolConfig

		mcpClients map[string]mcpclient.MCPClient
		mtx        sync.Mutex
	}
	LocalToolServiceKey string
)

var (
	_                   LocalToolService    = (*manager)(nil)
	_                   Manager             = (*manager)(nil)
	localToolServiceKey LocalToolServiceKey = "agentruntime.local_tool_service"
)

func (m *manager) GetTool(_ context.Context, toolName string) ai.Tool {
	tool := ai.LookupTool(toolName)
	if tool.Action() == nil {
		return nil
	}

	return tool
}

func (m *manager) GetMCPTool(_ context.Context, serverName, toolName string) ai.Tool {
	if _, ok := m.mcpClients[serverName]; !ok {
		return nil
	}

	tool := ai.LookupTool(toolName)
	if tool.Action() == nil {
		return nil
	}
	return tool
}

func (m *manager) Close() {
	for _, client := range m.mcpClients {
		if err := client.Close(); err != nil {
			m.logger.Warn("failed to close MCP client", "err", err)
		}
	}
}

func registerLocalTool[In any, Out any](name string, description string, fn func(context.Context, In) (Out, error)) ai.Tool {
	return ai.DefineTool(
		name,
		description,
		func(ctx context.Context, input In) (Out, error) {
			out, err := fn(ctx, input)
			if err == nil {
				appendCallData(ctx, CallData{
					Name:      name,
					Arguments: input,
					Result:    out,
				})
			}
			return out, err
		},
	)
}

func WithLocalToolService(ctx context.Context, toolService LocalToolService) context.Context {
	return context.WithValue(ctx, localToolServiceKey, toolService)
}
