package mcp

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type ToolResult struct {
	Error  string          `json:"error,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

type MCPClientRegistry interface {
	GetMCPClient(ctx context.Context, serverName string) (mcpclient.MCPClient, error)
}

type DefaultMCPClientRegistry struct {
	Registry map[string]mcpclient.MCPClient
}

var mcpClientRegistryKey = "mcp_client.registry"

func (r *DefaultMCPClientRegistry) GetMCPClient(ctx context.Context, serverName string) (mcpclient.MCPClient, error) {
	return r.Registry[serverName], nil
}

func (r *ToolResult) String() string {
	if r.Error != "" {
		return r.Error
	}
	if r.Result != nil {
		return string(r.Result)
	}
	return ""
}

// DefineTool defines a tool function.
func DefineTool(serverName string, mcpTool mcp.Tool, cb func(ctx context.Context, input any, output *ToolResult) error) (ai.Tool, error) {
	metadata := make(map[string]any)
	metadata["type"] = "tool"
	metadata["name"] = mcpTool.Name
	metadata["description"] = mcpTool.Description

	schema, err := makeInputSchema(mcpTool.InputSchema)
	if err != nil {
		return nil, err
	}
	core.DefineActionWithInputSchema(
		"local",
		mcpTool.Name,
		"tool",
		metadata,
		schema,
		func(ctx context.Context, in any) (out *ToolResult, err error) {
			registry, ok := ctx.Value(mcpClientRegistryKey).(MCPClientRegistry)
			if !ok {
				return nil, errors.New("mcp client registry not found")
			}
			mcpClient, err := registry.GetMCPClient(ctx, serverName)
			if err != nil {
				return nil, err
			}
			if err = mcpClient.Ping(ctx); err != nil {
				return
			}

			req := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "tools/call",
				},
			}
			req.Params.Name = mcpTool.Name
			req.Params.Arguments = in

			var result *mcp.CallToolResult
			if result, err = mcpClient.CallTool(ctx, req); err != nil {
				return
			}

			out, err = processResult(result)
			if err != nil {
				return
			}

			if cb != nil {
				if err = cb(ctx, in, out); err != nil {
					return
				}
			}

			return out, nil
		},
	)

	return ai.LookupTool(mcpTool.Name), nil
}

func WithMCPClientRegistry(ctx context.Context, registry MCPClientRegistry) context.Context {
	return context.WithValue(ctx, mcpClientRegistryKey, registry)
}
