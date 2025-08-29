package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// QueryRewriter defines the interface for query rewriting strategies
type QueryRewriter interface {
	// Rewrite takes a query and returns one or more rewritten queries
	// The returned queries should be used for embedding and search
	Rewrite(ctx context.Context, query string) ([]string, error)
}

// NoOpQueryRewriter returns the original query unchanged
type NoOpQueryRewriter struct{}

func NewNoOpQueryRewriter() QueryRewriter {
	return &NoOpQueryRewriter{}
}

func (r *NoOpQueryRewriter) Rewrite(ctx context.Context, query string) ([]string, error) {
	return []string{query}, nil
}

// HyDERewriter implements Hypothetical Document Embeddings
// It generates a hypothetical answer to the query and uses both for search
type HyDERewriter struct {
	genkit *genkit.Genkit
	model  string
}

func NewHyDERewriter(genkit *genkit.Genkit, model string) QueryRewriter {
	if model == "" {
		model = "openai/gpt-5-mini"
	}
	return &HyDERewriter{
		genkit: genkit,
		model:  model,
	}
}

func (r *HyDERewriter) Rewrite(ctx context.Context, query string) ([]string, error) {
	prompt := fmt.Sprintf(`Given this question: "%s"

Write a comprehensive, factual answer that would typically be found in a knowledge base or documentation.
Be specific and include relevant details, but keep it concise (2-3 paragraphs).
Focus on information that directly answers the question.

Answer:`, query)

	response, err := genkit.Generate(ctx, r.genkit,
		ai.WithModelName(r.model),
		ai.WithPrompt(prompt),
		ai.WithOutputFormat(ai.OutputFormatText),
	)
	if err != nil {
		// On error, return original query
		return []string{query}, nil
	}

	// Return both original query and hypothetical answer
	return []string{query, response.Text()}, nil
}

// QueryExpansionRewriter expands queries with synonyms and related terms
type QueryExpansionRewriter struct {
	genkit *genkit.Genkit
	model  string
}

func NewQueryExpansionRewriter(genkit *genkit.Genkit, model string) QueryRewriter {
	if model == "" {
		model = "openai/gpt-5-mini"
	}
	return &QueryExpansionRewriter{
		genkit: genkit,
		model:  model,
	}
}

type expansionResult struct {
	Synonyms      []string `json:"synonyms"`
	RelatedTerms  []string `json:"related_terms"`
	ExpandedQuery string   `json:"expanded_query"`
}

func (r *QueryExpansionRewriter) Rewrite(ctx context.Context, query string) ([]string, error) {
	prompt := fmt.Sprintf(`Given this search query: "%s"

Provide query expansions to improve search results. Return a JSON object with:
1. synonyms: Alternative words or phrases with similar meaning
2. related_terms: Closely related concepts that might appear in relevant documents
3. expanded_query: A reformulated query incorporating key expansions

Keep expansions relevant and avoid overly broad terms.

Example format:
{
  "synonyms": ["alternative1", "alternative2"],
  "related_terms": ["related1", "related2"],
  "expanded_query": "original query with key alternatives and related concepts"
}

JSON:`, query)

	var result expansionResult
	response, err := genkit.Generate(ctx, r.genkit,
		ai.WithModelName(r.model),
		ai.WithPrompt(prompt),
		ai.WithOutputType(&result),
	)
	if err != nil {
		// On error, return original query
		return []string{query}, nil
	}

	// Parse the output
	if err := response.Output(&result); err != nil {
		// Try manual parsing as fallback
		var manualResult expansionResult
		if err := json.Unmarshal([]byte(response.Text()), &manualResult); err == nil {
			result = manualResult
		} else {
			// On error, return original query
			return []string{query}, nil
		}
	}

	// Build expanded queries
	queries := []string{query}

	// Add expanded query if available
	if result.ExpandedQuery != "" && result.ExpandedQuery != query {
		queries = append(queries, result.ExpandedQuery)
	}

	// Add a query with synonyms if available
	if len(result.Synonyms) > 0 {
		synonymQuery := query + " " + strings.Join(result.Synonyms, " ")
		queries = append(queries, synonymQuery)
	}

	return queries, nil
}

// MultiStrategyRewriter combines multiple rewriting strategies
type MultiStrategyRewriter struct {
	rewriters []QueryRewriter
}

func NewMultiStrategyRewriter(rewriters ...QueryRewriter) QueryRewriter {
	return &MultiStrategyRewriter{
		rewriters: rewriters,
	}
}

func (r *MultiStrategyRewriter) Rewrite(ctx context.Context, query string) ([]string, error) {
	// Use a map to track unique queries
	uniqueQueries := make(map[string]bool)
	var allQueries []string

	// Always include the original query
	uniqueQueries[query] = true
	allQueries = append(allQueries, query)

	// Apply each rewriter
	for _, rewriter := range r.rewriters {
		queries, err := rewriter.Rewrite(ctx, query)
		if err != nil {
			// Log error but continue with other rewriters
			continue
		}

		for _, q := range queries {
			q = strings.TrimSpace(q)
			if q != "" && !uniqueQueries[q] {
				uniqueQueries[q] = true
				allQueries = append(allQueries, q)
			}
		}
	}

	return allQueries, nil
}

// CreateQueryRewriter creates a query rewriter based on configuration
func CreateQueryRewriter(genkit *genkit.Genkit, strategy string, model string) QueryRewriter {
	switch strings.ToLower(strategy) {
	case "hyde":
		return NewHyDERewriter(genkit, model)
	case "expansion":
		return NewQueryExpansionRewriter(genkit, model)
	case "multi":
		return NewMultiStrategyRewriter(
			NewHyDERewriter(genkit, model),
			NewQueryExpansionRewriter(genkit, model),
		)
	case "none", "":
		return NewNoOpQueryRewriter()
	default:
		// Default to no-op for unknown strategies
		return NewNoOpQueryRewriter()
	}
}
