package engine

import (
	"context"
	"html/template"
	"reflect"
	"slices"
	"strings"

	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/errors"

	"github.com/firebase/genkit/go/ai"
)

type (
	EvaluatorResponse struct {
		Score      float32  `json:"score" jsonschema_description:"Score of the response. 0.0 is the worst, 1.0 is the best"`
		Reason     string   `json:"reason" jsonschema_description:"Reason of the score. It should be a short sentence."`
		Suggestion []string `json:"suggestion" jsonschema_description:"Suggestion to improve the response. It should be a short sentence."`
	}
	GenerateRequest struct {
		Vars                any
		PromptTmpl          string
		SystemPromptTmpl    string
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

func (e *engine) Generate(
	ctx context.Context,
	req *GenerateRequest,
	out any,
	opts ...ai.GenerateOption,
) (*ai.ModelResponse, error) {
	if out == nil {
		return nil, errors.New("output is nil")
	}
	isObjectOutput := reflect.TypeOf(out).Elem().Kind() != reflect.String

	if req.PromptTmpl != "" {
		opts = append(opts, ai.WithPromptFn(func(ctx context.Context, _ any) (string, error) {
			var prompt strings.Builder
			promptTmpl, err := template.New("").Funcs(funcMap()).Parse(req.PromptTmpl)
			if err != nil {
				return "", err
			}

			if err := promptTmpl.Execute(&prompt, req.Vars); err != nil {
				return "", err
			}
			return prompt.String(), nil
		}))
	}

	if req.SystemPromptTmpl != "" {
		opts = append(opts, ai.WithSystemFn(func(ctx context.Context, _ any) (string, error) {
			var systemPrompt strings.Builder
			systemPromptTmpl, err := template.New("").Funcs(funcMap()).Parse(req.SystemPromptTmpl)
			if err != nil {
				return "", err
			}
			if err := systemPromptTmpl.Execute(&systemPrompt, req.Vars); err != nil {
				return "", err
			}

			return systemPrompt.String(), nil
		}))
	}

	if isObjectOutput {
		opts = append(opts, ai.WithOutputType(out))
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

	if isObjectOutput {
		if err := resp.Output(out); err != nil {
			return nil, err
		}
	} else {
		reflect.ValueOf(out).Elem().SetString(resp.Text())
	}

	return resp, nil
}
