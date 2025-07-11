package engine

import (
	"context"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type (
	GenerateRequest struct {
		Model string
	}
)

func (e *Engine) Generate(
	ctx context.Context,
	req *GenerateRequest,
	opts ...ai.GenerateOption,
) (*ai.ModelResponse, error) {
	modelName := req.Model
	modelNamePieces := strings.SplitN(modelName, "/", 2)
	if len(modelNamePieces) == 1 {
		modelName = "openai/" + req.Model
	}
	opts = append(opts, ai.WithModelName(modelName))

	resp, err := genkit.Generate(ctx, e.genkit, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
