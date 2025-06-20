package engine

import (
	"context"
	"slices"
	"strings"

	"github.com/firebase/genkit/go/genkit"
	"github.com/pkg/errors"

	"github.com/firebase/genkit/go/ai"
)

type (
	EvaluatorResponse struct {
		Score      float32  `json:"score" jsonschema_description:"Score of the response. 0.0 is the worst, 1.0 is the best"`
		Reason     string   `json:"reason" jsonschema_description:"Reason of the score. It should be a short sentence."`
		Suggestion []string `json:"suggestion" jsonschema_description:"Suggestion to improve the response. It should be a short sentence."`
	}
	GenerateRequest struct {
		Model               string
		EvaluatorPromptTmpl string
		NumRetries          int
	}
)

func withoutPurposeOutput(history []*ai.Message) []*ai.Message {
	newHistory := slices.Clone(history)
	for _, hist := range newHistory {
		for i, c := range hist.Content {
			if c.Metadata != nil && c.Metadata["purpose"] == "output" {
				hist.Content = slices.Delete(hist.Content, i, i+1)
				break
			}
		}
	}

	return newHistory
}

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

	for i := 0; i < req.NumRetries; i++ {
		options := []ai.GenerateOption{
			ai.WithModelName(modelName),
			ai.WithPrompt("Please evaluate it."),
			ai.WithMessages(withoutPurposeOutput(resp.History())...),
			ai.WithMaxTurns(100),
		}
		if i == 0 {
			options = append(options, ai.WithSystem(req.EvaluatorPromptTmpl))
		}
		answer, evalRes, err := genkit.GenerateData[EvaluatorResponse](ctx, e.genkit, options...)
		if err != nil {
			return nil, err
		}
		if answer.Score >= 0.95 {
			break
		}
		e.logger.Info("retrying", "score", answer.Score, "reason", answer.Reason, "suggestion", answer.Suggestion)

		resp, err = genkit.Generate(
			ctx,
			e.genkit,
			ai.WithModelName(modelName),
			ai.WithPrompt("Please fix it."),
			ai.WithMessages(withoutPurposeOutput(evalRes.History())...),
			ai.WithOutputType(out),
			ai.WithMaxTurns(100),
		)
		if err != nil {
			return nil, err
		}
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
