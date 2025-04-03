package tool

import (
	"context"
	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/mylog"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"sync"
)

type (
	manager struct {
		logger *mylog.Logger
		config *config.RuntimeConfig

		mcpClients map[string]mcpclient.MCPClient
		mtx        sync.Mutex
	}
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

var (
	_ Manager = (*manager)(nil)
)

func RegisterLocalTool[In any, Out any](name string, description string, fn func(context.Context, In) (Out, error)) ai.Tool {
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
