package engine

import (
	"context"
	_ "embed"
	"encoding/json"
	"reflect"
	"strings"
	"text/template"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/entity"
	myerrors "github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/internal/sliceutils"
	"github.com/habiliai/agentruntime/tool"
	"github.com/pkg/errors"
)

var (
	//go:embed data/instructions/chat.md.tmpl
	chatInst     string
	chatInstTmpl = template.Must(template.New("chatInst").Funcs(funcMap()).Parse(chatInst))
)

type (
	Action struct {
		Name      string `json:"name"`
		Arguments any    `json:"arguments"`
		Result    any    `json:"result"`
	}
	Conversation struct {
		User    string   `json:"user"`
		Text    string   `json:"text"`
		Actions []Action `json:"actions,omitempty"`
	}

	AvailableAction struct {
		Action      string `json:"action"`
		Description string `json:"description"`
	}

	Participant struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}

	ThreadValues struct {
		Instruction  string
		Participants []Participant `json:"participants,omitempty"`
	}

	ChatInstValues struct {
		Agent               entity.Agent
		RecentConversations []Conversation
		Knowledge           []string
		AvailableActions    []AvailableAction
		MessageExamples     [][]entity.MessageExample
		Thread              ThreadValues
	}

	RunRequest struct {
		ThreadInstruction string         `json:"thread_instruction,omitempty"`
		History           []Conversation `json:"history"`
		Agent             entity.Agent   `json:"agent"`
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

func (s *engine) Run(
	ctx context.Context,
	req RunRequest,
	output any,
) (*RunResponse, error) {
	if output == nil {
		return nil, errors.Errorf("output is nil")
	} else if reflect.TypeOf(output).Kind() != reflect.Ptr {
		return nil, errors.Errorf("output is not a pointer")
	}
	agent := req.Agent

	// construct inst values
	instValues := ChatInstValues{
		Agent:               agent,
		MessageExamples:     sliceutils.RandomSampleN(agent.MessageExamples, 100),
		RecentConversations: sliceutils.Cut(req.History, -25, len(req.History)),
		AvailableActions:    make([]AvailableAction, 0, len(agent.Tools)),
		Thread: ThreadValues{
			Instruction:  req.ThreadInstruction,
			Participants: req.Participant,
		},
	}

	// build available actions
	tools := make([]ai.ToolRef, 0, len(agent.Tools))
	for _, tool := range agent.Tools {
		instValues.AvailableActions = append(instValues.AvailableActions, AvailableAction{
			Action:      tool.Name,
			Description: tool.Description,
		})

		toolNames := strings.SplitN(tool.Name, "/", 2)
		var v ai.Tool
		if len(toolNames) == 1 {
			v = s.toolManager.GetTool(tool.Name)
		} else {
			v = s.toolManager.GetMCPTool(toolNames[0], toolNames[1])
		}
		if v == nil {
			return nil, errors.Wrapf(myerrors.ErrInvalidConfig, "invalid tool name %s", tool.Name)
		}
		tools = append(tools, v)
	}

	ctx = tool.WithEmptyCallDataStore(ctx)
	var (
		err error
	)
	for i := 0; i < 3; i++ {
		_, err = s.Generate(
			ctx,
			&GenerateRequest{
				Vars:                instValues,
				PromptTmpl:          chatInst,
				Model:               agent.ModelName,
				SystemPromptTmpl:    agent.System,
				EvaluatorPromptTmpl: agent.Evaluator.Prompt,
				NumRetries:          agent.Evaluator.NumRetries,
			},
			output,
			ai.WithConfig(agent.ModelConfig),
			ai.WithTools(tools...),
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
