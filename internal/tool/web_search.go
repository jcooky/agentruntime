package tool

import (
	"context"
)

type (
	WebSearchToolRequest struct {
		Query string `json:"query"`
	}
	WebSearchToolResponse struct {
		Result string `json:"result"`
	}
)

func (m *manager) registerWebSearchTool() {
	registerLocalTool(
		m,
		"web_search",
		"This is dummy tool for web search",
		func(ctx context.Context, req struct {
			*WebSearchToolRequest
		}) (res struct {
			*WebSearchToolResponse
		}, err error) {
			return struct {
				*WebSearchToolResponse
			}{
				WebSearchToolResponse: &WebSearchToolResponse{
					Result: "This is dummy tool for web search",
				},
			}, nil
		},
	)
}
