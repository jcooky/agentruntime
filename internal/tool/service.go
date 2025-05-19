package tool

import (
	"context"
	"sync"

	"github.com/firebase/genkit/go/genkit"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/mylog"
	mcpclient "github.com/mark3labs/mcp-go/client"
)

type (
	manager struct {
		logger *mylog.Logger
		config *config.ToolConfig

		mcpClients map[string]mcpclient.MCPClient
		mtx        sync.Mutex
		genkit     *genkit.Genkit
	}
)

var (
	_ Manager = (*manager)(nil)
)

func (m *manager) GetTool(toolName string) ai.Tool {
	return genkit.LookupTool(m.genkit, toolName)
}

func (m *manager) GetMCPTool(serverName, toolName string) ai.Tool {
	if _, ok := m.mcpClients[serverName]; !ok {
		return nil
	}

	return genkit.LookupTool(m.genkit, toolName)
}

func (m *manager) Close() {
	for _, client := range m.mcpClients {
		if err := client.Close(); err != nil {
			return
		}
	}
}

func registerLocalTool[In any, Out any](m *manager, name string, description string, fn func(context.Context, In) (Out, error)) ai.Tool {
	tool := m.GetTool(name)
	if tool != nil {
		return tool
	}

	return genkit.DefineTool(
		m.genkit,
		name,
		description,
		func(ctx *ai.ToolContext, input In) (Out, error) {
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
