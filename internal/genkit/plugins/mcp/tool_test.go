package mcp_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/mcp"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

func TestMCPToolCall(t *testing.T) {
	ctx := context.TODO()
	g, err := genkit.Init(ctx)
	if err != nil {
		t.Fatalf("failed to create genkit: %v", err)
	}

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
	tool, err := mcp.DefineTool(g, c, listDirTool, nil)
	if err != nil {
		t.Fatalf("failed to define tool: %v", err)
	}

	t.Run("Run Tool", func(t *testing.T) {
		out, err := tool.RunRaw(ctx, json.RawMessage(`{"path":"./"}`))
		if err != nil {
			t.Fatalf("failed to run tool: %v", err)
		}

		output, ok := out.(map[string]any)
		if !ok {
			t.Fatalf("result is not a ToolResult")
		}

		t.Logf("result: %+v", output)
	})
}
