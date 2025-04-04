package mcp

import (
	"encoding/json"
	"github.com/invopop/jsonschema"
	"github.com/mark3labs/mcp-go/mcp"
	"strings"
)

func makeInputSchema(
	schema mcp.ToolInputSchema,
) (*jsonschema.Schema, error) {
	var inputSchema jsonschema.Schema

	schemaJson, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(schemaJson, &inputSchema); err != nil {
		return nil, err
	}

	return &inputSchema, nil
}

func toText(contents []mcp.Content) string {
	text := ""
	for _, c := range contents {
		if t, ok := c.(mcp.TextContent); ok {
			text += t.Text
		}
	}

	return strings.TrimSpace(text)
}

func processResult(result *mcp.CallToolResult) (*ToolResult, error) {
	if result == nil {
		return nil, nil
	}

	var (
		out ToolResult
		err error
	)
	if result.IsError {
		out.Error = toText(result.Content)
		return &out, nil
	} else {
		content := any(result.Content)
		if len(result.Content) == 1 {
			content = result.Content[0]
		}
		out.Result, err = json.Marshal(content)
		return &out, err
	}
}
