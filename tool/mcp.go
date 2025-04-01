package tool

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/plugins/mcp"
	mcpclient "github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
	"runtime/debug"
)

type RegisterMCPToolRequest struct {
	ServerName string
	Command    string
	Args       []string
	Env        map[string]string
}

func (m *manager) RegisterMCPTool(ctx context.Context, req RegisterMCPToolRequest) (err error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	var envs []string
	for key, val := range req.Env {
		envs = append(envs, fmt.Sprintf("%s=%s", key, val))
	}

	mcpClient, ok := m.mcpClients[req.ServerName]
	if !ok {
		mcpClient, err = mcpclient.NewStdioMCPClient(req.Command, envs, req.Args...)
		if err != nil {
			return fmt.Errorf("failed to create MCP client: %w", err)
		}

		initRequest := mcpgo.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
		if bi, ok := debug.ReadBuildInfo(); ok {
			initRequest.Params.ClientInfo = mcpgo.Implementation{
				Name:    bi.Main.Path,
				Version: bi.Main.Version,
			}
		}
		if _, err := mcpClient.Initialize(ctx, initRequest); err != nil {
			return errors.Wrapf(err, "failed to initialize MCP client")
		}

		m.mcpClients[req.ServerName] = mcpClient
	}

	listToolsResult, err := mcpClient.ListTools(ctx, mcpgo.ListToolsRequest{})
	if err != nil {
		return errors.Wrapf(err, "failed to list tools")
	}
	for _, tool := range listToolsResult.Tools {
		if mcp.LookupTool(req.ServerName, tool.Name).Action() != nil {
			m.logger.InfoContext(ctx, "tool already registered", "tool", tool.Name)
			continue
		}
		if _, err := mcp.DefineTool(mcpClient, req.ServerName, tool); err != nil {
			return errors.Wrapf(err, "failed to define tool")
		}
	}

	return nil
}

func (m *manager) GetMCPTools(_ context.Context, mcpServerName string) []ai.Tool {
	return mcp.LookupTools(mcpServerName)
}
