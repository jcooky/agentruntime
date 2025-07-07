package engine

import (
	"context"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/pkg/errors"
)

type (
	EvaluatorResponse struct {
		Score      float32  `json:"score" jsonschema_description:"Score of the response. 0.0 is the worst, 1.0 is the best"`
		Reason     string   `json:"reason" jsonschema_description:"Reason of the score. It should be a short sentence."`
		Suggestion []string `json:"suggestion" jsonschema_description:"Suggestion to improve the response. It should be a short sentence."`
	}
	GenerateRequest struct {
		Model string
	}
)

func (e *Engine) Generate(
	ctx context.Context,
	req *GenerateRequest,
	out any,
	opts ...ai.GenerateOption,
) (*ai.ModelResponse, error) {
	if out == nil {
		return nil, errors.New("output is nil")
	}
	switch v := out.(type) {
	case *string:
		opts = append(opts, ai.WithOutputFormat(ai.OutputFormatText))
	default:
		opts = append(opts, ai.WithOutputType(v))
	}

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

	switch v := out.(type) {
	case *string:
		*v = resp.Text()
	default:
		if err := resp.Output(v); err != nil {
			return nil, err
		}
	}

	return resp, nil
}
