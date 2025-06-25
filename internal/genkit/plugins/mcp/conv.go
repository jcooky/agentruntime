package mcp

import (
	"encoding/json"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/mark3labs/mcp-go/mcp"
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
