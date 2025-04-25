package tool

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/mcp"
	mcpclient "github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
)

type RegisterMCPToolRequest struct {
	ServerName string
	Command    string
	Args       []string
	Env        map[string]string
}

var (
	_ mcp.MCPClientRegistry = (*manager)(nil)
)

func (m *manager) GetMCPClient(ctx context.Context, serverName string) (mcpclient.MCPClient, error) {
	return m.mcpClients[serverName], nil
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
		c, err := mcpclient.NewStdioMCPClient(req.Command, envs, req.Args...)
		if err != nil {
			return fmt.Errorf("failed to create MCP client: %w", err)
		}

		go func(stderr io.Reader) {
			rd := bufio.NewReader(stderr)
			for {
				line, err := rd.ReadString('\n')
				if err != nil {
					if err == io.EOF || strings.Contains(err.Error(), "already closed") {
						return
					}
					m.logger.Error("failed to copy stderr", "err", err, "serverName", req.ServerName)
					return
				}
				m.logger.Warn("[MCP] "+strings.TrimSpace(line), "serverName", req.ServerName)
			}
		}(c.Stderr())

		initRequest := mcpgo.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
		if _, err := c.Initialize(ctx, initRequest); err != nil {
			return errors.Wrapf(err, "failed to initialize MCP client")
		}

		m.mcpClients[req.ServerName] = c
		mcpClient = c
	}

	listToolsResult, err := mcpClient.ListTools(ctx, mcpgo.ListToolsRequest{})
	if err != nil {
		return errors.Wrapf(err, "failed to list tools")
	}
	for _, tool := range listToolsResult.Tools {
		if ai.LookupTool(tool.Name).Action() != nil {
			m.logger.InfoContext(ctx, "tool already registered", "tool", tool.Name)
			continue
		}
		if _, err := mcp.DefineTool(req.ServerName, tool, func(ctx context.Context, in any, out *mcp.ToolResult) error {
			appendCallData(ctx, CallData{
				Name:      tool.Name,
				Arguments: in,
				Result:    out,
			})
			return nil
		}); err != nil {
			return errors.Wrapf(err, "failed to define tool")
		}
	}

	return nil
}

func (m *manager) GetMCPTools(ctx context.Context, mcpServerName string) []ai.Tool {
	client, ok := m.mcpClients[mcpServerName]
	if !ok {
		return nil
	}

	listToolsResult, err := client.ListTools(ctx, mcpgo.ListToolsRequest{})
	if err != nil {
		m.logger.Error("failed to list tools", "err", err)
		return nil
	}

	var tools []ai.Tool
	for _, tool := range listToolsResult.Tools {
		if t := ai.LookupTool(tool.Name); t != nil {
			tools = append(tools, t)
		}
	}

	return tools
}
