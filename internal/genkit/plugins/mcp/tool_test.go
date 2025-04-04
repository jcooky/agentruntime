package mcp_test

import (
	"context"
	"encoding/json"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/mcp"
	"testing"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

func TestMCPToolCall(t *testing.T) {
	ctx := context.TODO()

	c, err := mcpclient.NewStdioMCPClient("npx", []string{}, "-y", "@modelcontextprotocol/server-filesystem", ".")
	if err != nil {
		t.Fatalf("failed to create MCP client: %v", err)
	}
	defer c.Close()

	initRequest := mcpgo.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcpgo.Implementation{
		Name:    "example-client",
		Version: "1.0.0",
	}
	if _, err := c.Initialize(ctx, initRequest); err != nil {
		t.Fatalf("failed to initialize MCP client: %v", err)
	}

	listToolsRes, err := c.ListTools(ctx, mcpgo.ListToolsRequest{})
	if err != nil {
		t.Fatalf("failed to list tools: %v", err)
	}
	var listDirTool mcpgo.Tool
	for _, tool := range listToolsRes.Tools {
		if tool.Name == "list_directory" {
			listDirTool = tool
			break
		}
	}
	tool, err := mcp.DefineTool(c, listDirTool, nil)
	if err != nil {
		t.Fatalf("failed to define tool: %v", err)
	}

	t.Run("Run Tool", func(t *testing.T) {
		out, err := tool.Action().RunJSON(ctx, []byte(`{"path":"./"}`), nil)
		if err != nil {
			t.Fatalf("failed to run tool: %v", err)
		}

		var output mcp.ToolResult
		if err := json.Unmarshal(out, &output); err != nil {
			t.Fatalf("failed to unmarshal result: %v", err)
		}

		t.Logf("result: %+v", output)
	})
}
