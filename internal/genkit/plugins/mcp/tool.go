package mcp

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// DefineTool defines a tool function.
func DefineTool(g *genkit.Genkit, client client.MCPClient, mcpTool mcp.Tool, cb func(ctx *ai.ToolContext, input any, output *mcp.CallToolResult) error) (ai.Tool, error) {
	schema, err := makeInputSchema(mcpTool.InputSchema)
	if err != nil {
		return nil, err
	}

	tool := genkit.DefineToolWithInputSchema(
		g,
		mcpTool.Name,
		mcpTool.Description,
		schema,
		func(ctx *ai.ToolContext, in any) (out *mcp.CallToolResult, err error) {
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

			if out, err = client.CallTool(ctx, req); err != nil {
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
