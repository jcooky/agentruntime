package runner

import (
	"context"
	_ "embed"
	"encoding/json"
	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/entity"
	myerrors "github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/tool"
	"github.com/pkg/errors"
	"github.com/yukinagae/genkit-go-plugins/plugins/openai"
	"strings"
	"text/template"
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
		ThreadInstruction string         `json:"thread_instruction"`
		History           []Conversation `json:"history"`
		Agent             entity.Agent   `json:"agents"`
		Participant       []Participant  `json:"participants,omitempty"`
	}

	RunResponse struct {
		Content   string                `json:"content"`
		ToolCalls []RunResponseToolcall `json:"tool_calls"`
	}

	RunResponseToolcall struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
		Result    json.RawMessage `json:"result"`
	}
)

func (s *runner) Run(
	ctx context.Context,
	req RunRequest,
) (*RunResponse, error) {
	agent := req.Agent
	// construct inst values
	instValues := ChatInstValues{
		Agent:               agent,
		MessageExamples:     agent.MessageExamples,
		RecentConversations: req.History,
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
	case "gpt-4o":
		config = ai.GenerationCommonConfig{
			Temperature: 0.2,
			TopP:        0.5,
			TopK:        16,
		}
	default:
		return nil, errors.Errorf("unsupported model %s", agent.ModelName)
	}

	ctx = tool.WithEmptyCallDataStore(ctx)
	var (
		resp *ai.GenerateResponse
		err  error
	)
	for i := 0; i < 3; i++ {
		resp, err = ai.Generate(
			ctx,
			model,
			ai.WithCandidates(1),
			ai.WithSystemPrompt(agent.System),
			ai.WithTextPrompt(prompt),
			ai.WithConfig(config),
			ai.WithOutputFormat(ai.OutputFormatJSON),
			ai.WithOutputSchema(&Conversation{}),
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

	responseText := resp.Text()

	var conversation struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(responseText), &conversation); err != nil {
		s.logger.Debug("failed to unmarshal conversation", "responseText", responseText)
		return nil, errors.Wrapf(err, "failed to unmarshal conversation")
	}

	res := RunResponse{
		Content: conversation.Text,
	}

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
