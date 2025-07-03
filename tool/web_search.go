package tool

import "github.com/habiliai/agentruntime/entity"

type (
	WebSearchToolRequest struct {
		Query string `json:"query"`
	}
	WebSearchToolResponse struct {
		Result string `json:"result"`
	}
)

func (m *manager) registerWebSearchTool() error {
	return registerNativeTool(
		m,
		"web_search",
		"This is dummy tool for web search",
		&entity.NativeAgentSkill{
			Name:    "web_search",
			Details: "This is dummy tool for web search",
			Env:     map[string]any{},
		},
		func(ctx *Context, req struct {
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
