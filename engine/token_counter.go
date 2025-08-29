package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/anthropic"
	"github.com/pkg/errors"
)

func CountTokens(
	ctx context.Context,
	g *genkit.Genkit,
	provider string, // openai, anthropic, xai, etc.
	msgs []*ai.Message,
	docs []*ai.Document,
	toolDefs []ai.Tool,
) (int, error) {
	switch strings.ToLower(provider) {
	case "openai", "gpt", "gpt-4", "gpt-3.5", "gpt-5":
		return 0, errors.New("openai token counter is not supported")

	case "anthropic", "claude":
		// For Anthropic, we need to know the specific model
		// Default to claude-3-5-sonnet if not specified
		return anthropic.CountTokens(ctx, g, msgs, docs, toolDefs)

	default:
		return 0, fmt.Errorf("unsupported token counter provider: %s", provider)
	}
}
