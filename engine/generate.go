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
	GenerateRequest struct {
		Vars             any
		PromptTmpl       string
		SystemPromptTmpl string
		Model            string
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

	if isObjectOutput {
		if err := resp.UnmarshalOutput(out); err != nil {
			return nil, err
		}
	} else {
		reflect.ValueOf(out).Elem().SetString(resp.Text())
	}

	return resp, nil
}
