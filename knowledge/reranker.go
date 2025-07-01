package knowledge

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// Reranker interface for reranking retrieval results
type Reranker interface {
	// Rerank takes a query and candidate results, returns reranked results with scores
	Rerank(ctx context.Context, query string, candidates []*KnowledgeSearchResult, topK int) ([]*KnowledgeSearchResult, error)
}

// GenkitReranker implements Reranker using genkit LLM for relevance scoring
type GenkitReranker struct {
	genkit *genkit.Genkit
	model  string
}

// NewGenkitReranker creates a new reranker using genkit
func NewGenkitReranker(genkit *genkit.Genkit, model string) Reranker {
	// Add openai prefix if not present
	if model != "" && !strings.Contains(model, "/") {
		model = "openai/" + model
	}
	return &GenkitReranker{
		genkit: genkit,
		model:  model,
	}
}

// Rerank reranks the candidates based on relevance to the query
func (r *GenkitReranker) Rerank(ctx context.Context, query string, candidates []*KnowledgeSearchResult, topK int) ([]*KnowledgeSearchResult, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	// Limit topK to the number of candidates
	if topK > len(candidates) {
		topK = len(candidates)
	}

	results := make([]*KnowledgeSearchResult, 0, len(candidates))

	// Score each candidate
	for _, candidate := range candidates {
		score, err := r.scoreRelevance(ctx, query, candidate)
		if err != nil {
			// If scoring fails for one candidate, use a default low score
			score = 0.0
		}
		candidate.Score = float32(score)
		results = append(results, candidate)
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Return top K results
	if topK < len(results) {
		results = results[:topK]
	}

	return results, nil
}

// scoreRelevance uses LLM to score the relevance of a candidate to a query
func (r *GenkitReranker) scoreRelevance(ctx context.Context, query string, candidate *KnowledgeSearchResult) (float64, error) {
	prompt := fmt.Sprintf(`Rate the relevance of the document to the query on a scale from 0 to 10.
Query: %s

Respond with only a number between 0 and 10, where:
- 0 means completely irrelevant
- 5 means somewhat relevant
- 10 means highly relevant

Score:`, query)

	doc, err := candidate.Document.ToDoc()
	if err != nil {
		return 0, err
	}

	resp, err := genkit.Generate(ctx, r.genkit,
		ai.WithModelName(r.model),
		ai.WithPrompt(prompt),
		ai.WithOutputFormat(ai.OutputFormatText),
		ai.WithDocs(doc),
	)
	if err != nil {
		return 0, err
	}

	// Parse the score from the response
	var score float64
	_, err = fmt.Sscanf(resp.Text(), "%f", &score)
	if err != nil {
		return 0, err
	}

	// Normalize score to 0-1 range
	return score / 10.0, nil
}

// NoOpReranker is a reranker that returns candidates as-is (for when reranking is disabled)
type NoOpReranker struct{}

func NewNoOpReranker() Reranker {
	return &NoOpReranker{}
}

func (r *NoOpReranker) Rerank(ctx context.Context, query string, candidates []*KnowledgeSearchResult, topK int) ([]*KnowledgeSearchResult, error) {
	if topK > len(candidates) {
		topK = len(candidates)
	}

	results := make([]*KnowledgeSearchResult, topK)
	for i := 0; i < topK; i++ {
		results[i] = candidates[i]
	}

	return results, nil
}

// BatchGenkitReranker implements Reranker using genkit LLM with batch processing
type BatchGenkitReranker struct {
	genkit *genkit.Genkit
	model  string
}

// NewBatchGenkitReranker creates a new batch reranker using genkit
func NewBatchGenkitReranker(genkit *genkit.Genkit, model string) Reranker {
	// Add openai prefix if not present
	if model != "" && !strings.Contains(model, "/") {
		model = "openai/" + model
	}
	return &BatchGenkitReranker{
		genkit: genkit,
		model:  model,
	}
}

// Rerank reranks the candidates based on relevance to the query in a single batch
func (r *BatchGenkitReranker) Rerank(ctx context.Context, query string, candidates []*KnowledgeSearchResult, topK int) ([]*KnowledgeSearchResult, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	// Limit topK to the number of candidates
	if topK > len(candidates) {
		topK = len(candidates)
	}

	// Create batch prompt
	prompt := fmt.Sprintf(`Given the following query and documents, rate the relevance of each documents on a scale from 0 to 10.

Query: %s
`, query)

	prompt += `
Please respond with a JSON array of objects, each containing:
- "index": the candidate number (1-based)
- "score": relevance score (0-10)

Example format:
[{"index": 1, "score": 8.5}, {"index": 2, "score": 3.0}, ...]

Response:`

	// Call LLM once for all candidates
	type ScoreResult struct {
		Index int     `json:"index" jsonschema:"description=The candidate number (1-based)"`
		Score float64 `json:"score" jsonschema:"description=Relevance score (0-10)"`
	}

	docs := make([]*ai.Document, 0, len(candidates))
	for idx, candidate := range candidates {
		doc, err := candidate.Document.ToDoc()
		if err != nil {
			return nil, err
		}
		doc.Content = append(doc.Content, ai.NewTextPart(fmt.Sprintf("Index: %d", idx+1)))
		docs = append(docs, doc)
	}

	var scores []ScoreResult
	resp, err := genkit.Generate(ctx, r.genkit,
		ai.WithModelName(r.model),
		ai.WithPrompt(prompt),
		ai.WithOutputType(&scores),
		ai.WithDocs(docs...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate batch scores: %w", err)
	}

	if err := resp.Output(&scores); err != nil {
		return nil, fmt.Errorf("failed to parse batch scores: %w", err)
	}

	// Build results
	results := make([]*KnowledgeSearchResult, 0, len(candidates))
	scoreMap := make(map[int]float64)
	for _, score := range scores {
		if score.Index > 0 && score.Index <= len(candidates) {
			scoreMap[score.Index-1] = score.Score / 10.0 // Normalize to 0-1
		}
	}

	// Create results with scores
	for i, candidate := range candidates {
		score, exists := scoreMap[i]
		if !exists {
			score = 0.0 // Default score if missing
		}
		candidate.Score = float32(score)
		results = append(results, candidate)
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Return top K results
	if topK < len(results) {
		results = results[:topK]
	}

	return results, nil
}
