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
	"github.com/habiliai/agentruntime/internal/genkit/plugins/mcp"
	"github.com/habiliai/agentruntime/internal/sliceutils"
	"github.com/habiliai/agentruntime/tool"
	"github.com/pkg/errors"
	"github.com/yukinagae/genkit-go-plugins/plugins/openai"
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
		ToolCalls []RunResponseToolcall `json:"tool_calls"`
	}

	RunResponseToolcall struct {
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
		MessageExamples:     sliceutils.RandomSampleN(agent.MessageExamples, 7),
		RecentConversations: sliceutils.Cut(req.History, -25, len(req.History)),
		AvailableActions:    make([]AvailableAction, 0, len(agent.Tools)),
		Thread: ThreadValues{
			Instruction:  req.ThreadInstruction,
			Participants: req.Participant,
		},
	}

	// build available actions
	tools := make([]ai.Tool, 0, len(agent.Tools))
	for _, tool := range agent.Tools {
		instValues.AvailableActions = append(instValues.AvailableActions, AvailableAction{
			Action:      tool.Name,
			Description: tool.Description,
		})

		toolNames := strings.SplitN(tool.Name, "/", 2)
		var v ai.Tool
		if len(toolNames) == 1 {
			v = s.toolManager.GetTool(ctx, tool.Name)
		} else {
			v = s.toolManager.GetMCPTool(ctx, toolNames[0], toolNames[1])
		}
		if v == nil {
			return nil, errors.Wrapf(myerrors.ErrInvalidConfig, "invalid tool name %s", tool.Name)
		}
		tools = append(tools, v)
	}

	var promptBuf strings.Builder
	if err := chatInstTmpl.Execute(&promptBuf, instValues); err != nil {
		return nil, errors.Wrapf(err, "failed to execute template")
	}
	prompt := promptBuf.String()

	s.logger.Debug("call agent runtime's run", "prompt", prompt)

	model := openai.Model(agent.ModelName)

	var config any
	switch agent.ModelName {
	case "o1", "o3-mini":
		config = openai.GenerationReasoningConfig{
			ReasoningEffort: "high",
		}
	}

	ctx = tool.WithEmptyCallDataStore(ctx)
	ctx = mcp.WithMCPClientRegistry(ctx, s.toolManager)
	ctx = tool.WithLocalToolService(ctx, s.toolManager)
	var (
		responseText string
		err          error
		resp         *ai.GenerateResponse
		opts         []ai.GenerateOption
		format       = ai.OutputFormatText
	)
	if reflect.TypeOf(output).Elem().Kind() != reflect.String {
		format = ai.OutputFormatJSON
		opts = append(opts,
			ai.WithOutputFormat(format),
			ai.WithOutputSchema(output),
		)
	}
	opts = append(opts,
		ai.WithCandidates(1),
		ai.WithSystemPrompt(agent.System),
		ai.WithTextPrompt(prompt),
		ai.WithConfig(config),
		ai.WithTools(tools...),
	)

	for i := 0; i < 3; i++ {
		resp, err = ai.Generate(
			ctx,
			model,
			opts...,
		)
		if err != nil {
			s.logger.Warn("failed to generate", "err", err)
		} else {
			responseText = resp.Text()
			break
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate")
	}

	if format == ai.OutputFormatJSON {
		if err := json.Unmarshal([]byte(responseText), output); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal response")
		}
	} else {
		o, ok := output.(*string)
		if !ok {
			return nil, errors.Errorf("output is not a string pointer")
		}
		*o = responseText
	}

	var res RunResponse
	toolCallData := tool.GetCallData(ctx)
	for _, data := range toolCallData {
		tc := RunResponseToolcall{
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
