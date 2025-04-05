package runtime

import (
	"context"
	_ "embed"
	"encoding/json"
	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/entity"
	myerrors "github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/thread"
	"github.com/habiliai/agentruntime/tool"
	"github.com/mokiat/gog"
	"github.com/pkg/errors"
	"github.com/yukinagae/genkit-go-plugins/plugins/openai"
	"golang.org/x/sync/errgroup"
	"io"
	"slices"
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
)

func (s *service) Run(
	ctx context.Context,
	threadId uint,
	agents []entity.Agent,
) error {
	thr, err := s.threadManagerClient.GetThread(ctx, &thread.GetThreadRequest{
		ThreadId: uint32(threadId),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to get thread")
	}

	var messages []*thread.Message
	{
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		messagesStream, err := s.threadManagerClient.GetMessages(ctx, &thread.GetMessagesRequest{
			ThreadId: uint32(threadId),
		})
		if err != nil {
			return errors.Wrapf(err, "failed to get messages")
		}

		for {
			resp, err := messagesStream.Recv()
			if err == io.EOF {
				break
			} else if err != nil {
				return errors.Wrapf(err, "failed to receive messages")
			}

			messages = append(messages, resp.Messages...)
		}
	}

	slices.SortStableFunc(messages, func(a, b *thread.Message) int {
		if a.CreatedAt.AsTime().Before(b.CreatedAt.AsTime()) {
			return -1
		} else if a.CreatedAt.AsTime().After(b.CreatedAt.AsTime()) {
			return 1
		} else {
			return 0
		}
	})

	var eg errgroup.Group
	for _, agent := range agents {
		eg.Go(func() error {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			// construct inst values
			instValues := ChatInstValues{
				Agent:               agent,
				MessageExamples:     agent.MessageExamples,
				RecentConversations: make([]Conversation, 0, len(messages)),
				AvailableActions:    make([]AvailableAction, 0, len(agent.Tools)),
				Thread: ThreadValues{
					Instruction: thr.Instruction,
				},
			}

			// build recent conversations
			for _, msg := range messages {
				instValues.RecentConversations = append(instValues.RecentConversations, Conversation{
					User: msg.Sender,
					Text: msg.Content,
					Actions: gog.Map(msg.ToolCalls, func(tc *thread.Message_ToolCall) Action {
						return Action{
							Name:      tc.Name,
							Arguments: tc.Arguments,
							Result:    tc.Result,
						}
					}),
				})
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
					return errors.Wrapf(myerrors.ErrInvalidConfig, "invalid tool name %s", tool.Name)
				}
				tools = append(tools, v)
			}

			var promptBuf strings.Builder
			if err := chatInstTmpl.Execute(&promptBuf, instValues); err != nil {
				return errors.Wrapf(err, "failed to execute template")
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
				return errors.Errorf("unsupported model %s", agent.ModelName)
			}

			ctx = tool.WithEmptyCallDataStore(ctx)
			resp, err := ai.Generate(
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
				return errors.Wrapf(err, "failed to generate")
			}

			responseText := resp.Text()

			var conversation struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal([]byte(responseText), &conversation); err != nil {
				s.logger.Debug("failed to unmarshal conversation", "responseText", responseText)
				return errors.Wrapf(err, "failed to unmarshal conversation")
			}

			req := &thread.AddMessageRequest{
				ThreadId: uint32(threadId),
				Sender:   agent.Name,
				Content:  conversation.Text,
			}

			toolCallData := tool.GetCallData(ctx)
			for _, data := range toolCallData {
				tc := thread.Message_ToolCall{
					Name: data.Name,
				}

				if v, err := json.Marshal(data.Arguments); err != nil {
					return errors.Wrapf(err, "failed to marshal tool call arguments")
				} else {
					tc.Arguments = string(v)
				}

				if v, err := json.Marshal(data.Result); err != nil {
					return errors.Wrapf(err, "failed to marshal tool call result")
				} else {
					tc.Result = string(v)
				}

				req.ToolCalls = append(req.ToolCalls, &tc)
			}

			if _, err := s.threadManagerClient.AddMessage(ctx, req); err != nil {
				return errors.Wrapf(err, "failed to add message")
			}

			return nil
		})
	}

	return eg.Wait()
}
