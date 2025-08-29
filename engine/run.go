package engine

import (
	"context"
	_ "embed"
	"encoding/json"
	"math"
	"text/template"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/sliceutils"
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
		Files        []File        `json:"files,omitempty"`
	}

	ChatPromptValues struct {
		Agent               entity.Agent
		RecentConversations []Conversation
		AvailableActions    []AvailableAction
		MessageExamples     [][]entity.MessageExample
		Thread              Thread
		Tools               []ai.Tool
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

	promptValues, err := s.BuildPromptValues(ctx, agent, req, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build prompt values")
	}

	// Use conversation summarizer if available
	if s.conversationSummarizer != nil && len(req.History) > 0 {
		result, err := s.conversationSummarizer.ProcessConversationHistory(ctx, promptValues)
		if err != nil {
			return nil, err
		}

		req.History = result.RecentConversations
		promptValues, err = s.BuildPromptValues(ctx, agent, req, result.Summary)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to build prompt values")
		}
	} else {
		// Fall back to simple truncation when summarizer is not available
		recentConversations := sliceutils.Cut(req.History, -200, len(req.History))
		promptValues.RecentConversations = recentConversations
	}

	ctx = tool.WithEmptyCallDataStore(ctx)
	var res RunResponse
	res.ModelResponse, err = genkit.Generate(
		ctx,
		s.genkit,
		ai.WithModelName(agent.ModelName),
		ai.WithSystem(promptValues.System),
		ai.WithMessagesFn(func(ctx context.Context, _ any) ([]*ai.Message, error) {
			return convertToMessages(promptValues)
		}),
		ai.WithConfig(agent.ModelConfig),
		ai.WithTools(lo.Map(promptValues.Tools, func(t ai.Tool, _ int) ai.ToolRef {
			return t
		})...),
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
