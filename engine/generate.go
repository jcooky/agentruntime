package engine

import (
	"context"
	"html/template"
	"reflect"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/pkg/errors"
	"github.com/yukinagae/genkit-go-plugins/plugins/openai"
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

func (e *engine) Generate(
	ctx context.Context,
	req GenerateRequest,
	out any,
	opts ...ai.GenerateOption,
) (*ai.GenerateResponse, error) {
	if out == nil {
		return nil, errors.New("output is nil")
	}
	isObjectOutput := reflect.TypeOf(out).Elem().Kind() != reflect.String

	if req.PromptTmpl != "" {
		var prompt strings.Builder
		promptTmpl, err := template.New("").Funcs(funcMap()).Parse(req.PromptTmpl)
		if err != nil {
			return nil, err
		}

		if err := promptTmpl.Execute(&prompt, req.Vars); err != nil {
			return nil, err
		}
		opts = append(opts, ai.WithTextPrompt(prompt.String()))
	}

	if req.SystemPromptTmpl != "" {
		var systemPrompt strings.Builder
		systemPromptTmpl, err := template.New("").Funcs(funcMap()).Parse(req.SystemPromptTmpl)
		if err != nil {
			return nil, err
		}
		if err := systemPromptTmpl.Execute(&systemPrompt, req.Vars); err != nil {
			return nil, err
		}

		opts = append(opts, ai.WithSystemPrompt(systemPrompt.String()))
	}

	if isObjectOutput {
		opts = append(opts, ai.WithOutputSchema(out), ai.WithOutputFormat(ai.OutputFormatJSON))
	}

	model := openai.Model(req.Model)
	resp, err := ai.Generate(ctx, model, opts...)
	if err != nil {
		return nil, err
	}

	for i := 0; i < req.NumRetries; i++ {
		var answer EvaluatorResponse
		options := []ai.GenerateOption{
			ai.WithTextPrompt("Please evaluate it."),
			ai.WithHistory(resp.History()...),
			ai.WithOutputFormat(ai.OutputFormatJSON),
			ai.WithOutputSchema(&answer),
		}
		if i == 0 {
			options = append(options, ai.WithSystemPrompt(req.EvaluatorPromptTmpl))
		}
		evalRes, err := ai.Generate(ctx, model, options...)
		if err != nil {
			return nil, err
		}
		if err := evalRes.UnmarshalOutput(&answer); err != nil {
			return nil, err
		}
		if answer.Score >= 0.95 {
			break
		}
		resp, err = ai.Generate(ctx, model,
			ai.WithTextPrompt("Please fix it."),
			ai.WithHistory(evalRes.History()...),
			ai.WithOutputFormat(ai.OutputFormatJSON),
			ai.WithOutputSchema(out),
		)
		if err != nil {
			return nil, err
		}
	}

	if isObjectOutput {
		if err := resp.UnmarshalOutput(out); err != nil {
			return nil, err
		}
	} else {
		reflect.ValueOf(out).Elem().SetString(resp.Text())
	}

	return resp, nil
}
