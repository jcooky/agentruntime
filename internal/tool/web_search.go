package tool

import (
	"context"

	"github.com/habiliai/agentruntime/errors"
	g "github.com/serpapi/google-search-results-golang"
)

type (
	WebSearchRequest struct {
		Query string `json:"query" jsonschema_description:"The search query for the web search"`
		Type  string `json:"type" jsonschema_description:"The type of search to perform, can be 'google' only" jsonschema_enum:"google"`
	}
)

func (m *manager) Search(ctx context.Context, req *WebSearchRequest) ([]any, error) {
	query := g.NewSearch("google_light", map[string]string{
		"q": req.Query,
	}, m.config.SerpApiKey)

	search, err := query.GetJSON()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to perform web search")
	}
	results := search["organic_results"].([]interface{})
	return results, nil
}

func (m *manager) registerWebSearchTool() {
	registerLocalTool(
		m,
		"web_search",
		"Use when you need to search the web or find information online.",
		func(ctx context.Context, in struct {
			*WebSearchRequest
		}) (res []any, err error) {
			res, err = m.Search(ctx, in.WebSearchRequest)
			return
		},
	)
}
