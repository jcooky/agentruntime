package engine

import (
	"context"
	_ "embed"
	"encoding/json"
	"reflect"
	"text/template"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/tool"
	"github.com/pkg/errors"
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
	}

	RunRequest struct {
		ThreadInstruction string         `json:"thread_instruction,omitempty"`
		History           []Conversation `json:"history"`
		Participant       []Participant  `json:"participants,omitempty"`
	}

	RunResponse struct {
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
	output any,
) (*RunResponse, error) {
	if output == nil {
		return nil, errors.Errorf("output is nil")
	} else if reflect.TypeOf(output).Kind() != reflect.Ptr {
		return nil, errors.Errorf("output is not a pointer")
	}

	promptValues, err := s.BuildPromptValues(ctx, agent, req.History, Thread{
		Instruction:  req.ThreadInstruction,
		Participants: req.Participant,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build prompt values")
	}

	ctx = tool.WithEmptyCallDataStore(ctx)
	for i := 0; i < 3; i++ {
		_, err = s.Generate(
			ctx,
			&GenerateRequest{
				Model:               agent.ModelName,
				EvaluatorPromptTmpl: agent.Evaluator.Prompt,
				NumRetries:          agent.Evaluator.NumRetries,
			},
			output,
			ai.WithSystem(agent.System),
			ai.WithPromptFn(GetPromptFn(promptValues)),
			ai.WithConfig(agent.ModelConfig),
			ai.WithTools(promptValues.Tools...),
		)
		if err != nil {
			s.logger.Warn("failed to generate", "err", err)
		} else {
			break
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate")
	}

	var res RunResponse
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
