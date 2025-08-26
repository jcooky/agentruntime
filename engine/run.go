package engine

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
	"text/template"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/tool"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

const (
	defaultMaxTurns = math.MaxInt
)

var (
	//go:embed data/instructions/chat.md.tmpl
	chatInst     string
	chatInstTmpl *template.Template = template.Must(template.New("").Funcs(funcMap()).Parse(chatInst))
)

type (
	Action struct {
		Name      string `json:"name"`
		Arguments any    `json:"arguments"`
		Result    any    `json:"result"`
	}

	File struct {
		ContentType string `json:"content_type" jsonschema:"description=The MIME type of the file. accept only image/*, application/pdf"`
		Data        string `json:"data" jsonschema:"description=Base64 encoded data"`
		Filename    string `json:"filename"`
	}

	Conversation struct {
		User    string   `json:"user,omitempty"`
		Text    string   `json:"text,omitempty"`
		Actions []Action `json:"actions,omitempty"`
	}

	AvailableAction struct {
		Action      string `json:"action"`
		Description string `json:"description"`
	}

	Participant struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Role        string `json:"role"`
	}

	Thread struct {
		Instruction  string
		Participants []Participant `json:"participants,omitempty"`
	}

	ChatPromptValues struct {
		Agent               entity.Agent
		RecentConversations []Conversation
		AvailableActions    []AvailableAction
		MessageExamples     [][]entity.MessageExample
		Thread              Thread
		Tools               []ai.ToolRef
		System              string
		UserInfo            *UserInfo
	}

	RunRequest struct {
		ThreadInstruction string         `json:"thread_instruction,omitempty"`
		History           []Conversation `json:"history"`
		Participant       []Participant  `json:"participants,omitempty"`
		Files             []File         `json:"files"`
		UserInfo          *UserInfo      `json:"user_info"`
	}

	UserInfo struct {
		FullName string `json:"full_name"`
		Username string `json:"username,omitempty"`
		Location string `json:"location,omitempty"`
		Company  string `json:"company,omitempty"`
		Bio      string `json:"bio,omitempty"`
	}

	RunResponse struct {
		*ai.ModelResponse
		ToolCalls []ToolCall `json:"tool_calls"`
	}

	ToolCall struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
		Result    json.RawMessage `json:"result"`
	}
)

func (s *Engine) Run(
	ctx context.Context,
	agent entity.Agent,
	req RunRequest,
	streamCallback ai.ModelStreamCallback,
) (*RunResponse, error) {

	promptValues, err := s.BuildPromptValues(ctx, agent, req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build prompt values")
	}

	promptFn := GetPromptFn(promptValues)

	ctx = tool.WithEmptyCallDataStore(ctx)
	var res RunResponse
	res.ModelResponse, err = s.Generate(
		ctx,
		&GenerateRequest{
			Model: agent.ModelName,
		},
		ai.WithSystem(promptValues.System),
		ai.WithMessagesFn(func(ctx context.Context, _ any) ([]*ai.Message, error) {
			prompt, err := promptFn(ctx, nil)
			if err != nil {
				return nil, err
			}
			return []*ai.Message{
				{
					Role: ai.RoleUser,
					Content: slices.Concat(
						[]*ai.Part{
							ai.NewTextPart(prompt),
						},
						lo.Map(req.Files, func(f File, _ int) *ai.Part {
							return ai.NewMediaPart(f.ContentType, f.Data)
						}),
						[]*ai.Part{
							ai.NewTextPart(
								fmt.Sprintf(
									"<documents>Attached files:\n%s\n</documents>",
									strings.Join(lo.Map(req.Files, func(f File, i int) string {
										return fmt.Sprintf("%d. filename:'%s', content_type:'%s', data_length:%d", i+1, f.Filename, f.ContentType, len(f.Data))
									}), "\n"),
								),
							),
						},
					),
				},
			}, nil
		}),
		ai.WithConfig(agent.ModelConfig),
		ai.WithTools(promptValues.Tools...),
		ai.WithStreaming(streamCallback),
		ai.WithMaxTurns(defaultMaxTurns),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate response")
	}

	toolCallData := tool.GetCallData(ctx)
	for _, data := range toolCallData {
		tc := ToolCall{
			Name: data.Name,
		}

		if v, err := json.Marshal(data.Arguments); err != nil {
			return nil, errors.Wrapf(err, "failed to marshal tool call arguments")
		} else {
			tc.Arguments = v
		}

		if v, err := json.Marshal(data.Result); err != nil {
			return nil, errors.Wrapf(err, "failed to marshal tool call result")
		} else {
			tc.Result = v
		}

		res.ToolCalls = append(res.ToolCalls, tc)
	}

	return &res, nil
}
