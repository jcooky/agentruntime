package mcp

import (
	"encoding/json"
	"github.com/firebase/genkit/go/genkit"

	"github.com/firebase/genkit/go/ai"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type ToolResult struct {
	Error  string          `json:"error,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
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
func DefineTool(g *genkit.Genkit, client mcpclient.MCPClient, mcpTool mcp.Tool, cb func(ctx *ai.ToolContext, input any, output *ToolResult) error) (ai.Tool, error) {
	schema, err := makeInputSchema(mcpTool.InputSchema)
	if err != nil {
		return nil, err
	}

	tool := genkit.DefineToolWithInputSchema(
		g,
		mcpTool.Name,
		mcpTool.Description,
		schema,
		func(ctx *ai.ToolContext, in any) (out *ToolResult, err error) {
			if err = client.Ping(ctx); err != nil {
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
			if result, err = client.CallTool(ctx, req); err != nil {
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

	return tool, nil
}
