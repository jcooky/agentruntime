package memory

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/pkg/errors"
)

const keyPromptTemplate = `Generate memory key: category_subcategory_detail format

Prefixes:
- user_: Personal info (name, location, job, preferences, goals)
- project_: Work/project related  
- decision_: Important choices or agreements
- conversation_: Discussion context

Examples:
user_name_full, user_location_city, user_preference_coffee, user_goal_fitness
project_name_current, decision_architecture_2024

Rules: lowercase_with_underscores, specific, descriptive, unique
{{- if .ExistingKeys}}

Existing keys (avoid duplicates):
{{- range .ExistingKeys}}
{{.}}
{{- end}}
{{- end}}

Input: {{.Input}}
{{- if .Tags}}
Context: {{.TagsStr}}
{{- end}}
{{- if .CustomPrompt}}
{{.CustomPrompt}}
{{- end}}`

const tagsPromptTemplate = `Generate 1-3 categorization tags from these common types:

personal, preferences, work, goals, decisions, projects, relationships, skills, health, hobbies

Examples:
"I like coffee" → ["personal", "preferences"]  
"We chose React" → ["work", "decisions"]
"Learning Python" → ["work", "skills"]
{{- if .ExistingTags}}

Existing tags (reuse when appropriate): {{range $i, $tag := .ExistingTags}}{{if $i}}, {{end}}{{$tag}}{{end}}
{{- end}}

Input: {{.Input}}
{{- if .CustomPrompt}}
{{.CustomPrompt}}
{{- end}}`

type KeyTemplateData struct {
	Input        string
	Tags         []string
	TagsStr      string
	ExistingKeys []string
	CustomPrompt string
}

type TagsTemplateData struct {
	Input        string
	ExistingTags []string
	CustomPrompt string
}

var (
	keyTmpl  *template.Template
	tagsTmpl *template.Template
)

func init() {
	var err error
	keyTmpl, err = template.New("keyPrompt").Parse(keyPromptTemplate)
	if err != nil {
		panic(fmt.Sprintf("failed to parse key template: %v", err))
	}

	tagsTmpl, err = template.New("tagsPrompt").Parse(tagsPromptTemplate)
	if err != nil {
		panic(fmt.Sprintf("failed to parse tags template: %v", err))
	}
}

func (s *service) GenerateKey(ctx context.Context, input string, tags []string, prompt string, existingKeys []string) (string, error) {
	model, err := getModelForMini(s.genkit)
	if err != nil {
		return "", err
	}

	var output struct {
		Key string `json:"key" jsonschema:"required,description=Unique identifier using format: category_subcategory_detail (e.g. user_name_full, user_preference_coffee, project_requirements_2024)"`
	}

	data := KeyTemplateData{
		Input:        input,
		Tags:         tags,
		TagsStr:      strings.Join(tags, ", "),
		ExistingKeys: existingKeys,
		CustomPrompt: prompt,
	}

	var promptBuffer bytes.Buffer
	if err := keyTmpl.Execute(&promptBuffer, data); err != nil {
		return "", err
	}
	finalPrompt := strings.TrimSpace(promptBuffer.String())

	response, err := genkit.Generate(ctx, s.genkit, ai.WithModel(model), ai.WithPrompt(finalPrompt), ai.WithOutputType(&output))
	if err != nil {
		return "", err
	}

	if err := response.Output(&output); err != nil {
		return "", err
	}

	return output.Key, nil
}

func (s *service) GenerateTags(ctx context.Context, input string, prompt string, existingTags []string) ([]string, error) {

	model, err := getModelForMini(s.genkit)
	if err != nil {
		return nil, err
	}

	var output struct {
		Tags []string `json:"tags,omitempty" jsonschema:"description=Optional categorization tags (e.g. ['personal', 'preferences'], ['work', 'decisions'], ['goals'])"`
	}

	data := TagsTemplateData{
		Input:        input,
		ExistingTags: existingTags,
		CustomPrompt: prompt,
	}

	var promptBuffer bytes.Buffer
	if err := tagsTmpl.Execute(&promptBuffer, data); err != nil {
		return nil, err
	}
	finalPrompt := strings.TrimSpace(promptBuffer.String())

	response, err := genkit.Generate(ctx, s.genkit, ai.WithModel(model), ai.WithPrompt(finalPrompt), ai.WithOutputType(&output))
	if err != nil {
		return nil, err
	}

	if err := response.Output(&output); err != nil {
		return nil, err
	}

	return output.Tags, nil
}

func getModelForMini(g *genkit.Genkit) (ai.Model, error) {
	model := genkit.LookupModel(g, "openai", "gpt-4o-mini")
	if model != nil {
		return model, nil
	}

	model = genkit.LookupModel(g, "anthropic", "claude-3.5-haiku")
	if model != nil {
		return model, nil
	}

	return nil, errors.New("no model found. Please OPENAI_API_KEY and ANTHROPIC_API_KEY are set")
}
