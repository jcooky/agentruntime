package tool

import (
	"context"

	"github.com/habiliai/agentruntime/knowledge"
)

func (m *manager) registerKnowledgeSearchTool(knowledgeService knowledge.Service) {
	registerLocalTool(
		m,
		"knowledge_search",
		`Search through the external knowledge base to find relevant information, documents, and context.

This tool provides semantic search capabilities across stored knowledge including:
- Previously saved documents, notes, and information
- Context from past conversations and interactions
- Domain-specific knowledge and reference materials
- User preferences, patterns, and historical data

When to use this tool:
- When asked about information that might be stored in the knowledge base
- To retrieve context from previous conversations or saved documents
- When you need background information before answering a question
- To verify facts or find supporting evidence from stored knowledge
- When the user references something from the past ("as we discussed", "like last time", etc.)

How to use:
- Provide a clear, specific search query as input
- Use keywords and phrases that describe what you're looking for
- Consider multiple search attempts with different phrasings if initial results are insufficient
- The tool returns relevant excerpts with metadata about the source

Response structure:
- Results: Array of matching knowledge items with content and metadata
- Count: Number of results returned (may be less than limit if fewer matches found)
- Error: Error message if the search fails (e.g., connection issues, invalid query)

Error handling:
- If an error occurs, the Error field will contain the error message
- Common errors: database connection issues, invalid query format, service unavailable
- When an error occurs, try rephrasing the query or waiting before retry
- The tool will still return a response (not throw an error) to allow graceful handling

The search uses semantic similarity, so exact keyword matches are not required. Results are ranked by relevance and include context about when and where the information was stored.`,
		func(ctx context.Context, input struct {
			Query string `json:"query" jsonschema:"description=The search query to find relevant information"`
			Limit int    `json:"limit,omitempty" jsonschema:"description=The maximum number of results to return,default=10"`
		}) (reply struct {
			Results []*knowledge.KnowledgeSearchResult `json:"results" jsonschema:"description=List of search results with relevant knowledge"`
			Count   int                                `json:"count" jsonschema:"description=Number of results returned"`
			Error   string                             `json:"error,omitempty" jsonschema:"description=Error message if the search fails"`
		}, err error) {
			// Set default limit if not provided
			limit := input.Limit
			if limit <= 0 {
				limit = 10
			}

			// Retrieve relevant knowledge
			reply.Results, err = knowledgeService.RetrieveRelevantKnowledge(ctx, input.Query, limit)
			if err != nil {
				reply.Error = err.Error()
				return reply, nil
			}

			// Clean up embedding data to reduce response size
			for _, res := range reply.Results {
				res.EmbeddingText = ""
				res.Embeddings = nil
			}

			reply.Count = len(reply.Results)
			return
		},
	)
}
