package runtime

import (
	"context"
	_ "embed"
	"encoding/json"
	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/db"
	"github.com/habiliai/agentruntime/internal/myctx"
	"github.com/habiliai/agentruntime/tool"
	"github.com/mokiat/gog"
	"github.com/pkg/errors"
	"github.com/yukinagae/genkit-go-plugins/plugins/openai"
	"golang.org/x/sync/errgroup"
	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
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
	Conversation struct {
		User    string   `json:"user"`
		Text    string   `json:"text"`
		Actions []string `json:"actions"`
	}

	AvailableAction struct {
		Action      string `json:"action"`
		Description string `json:"description"`
	}

	ChatInstValues struct {
		Agent entity.Agent

		RecentConversations []Conversation
		Knowledge           []string
		AvailableActions    []AvailableAction
		MessageExamples     [][]entity.MessageExample
		Thread              *entity.Thread
	}
)

func (s *service) Run(
	ctx context.Context,
	threadId uint,
	agentIds []uint,
) error {
	ctx, sess := db.OpenSession(ctx, s.db)

	var thread entity.Thread
	if err := sess.First(&thread, threadId).Error; err != nil {
		return errors.Wrapf(err, "failed to find thread")
	}

	var agents []entity.Agent
	if err := sess.Preload(clause.Associations).Find(&agents, "id in ?", agentIds).Error; err != nil {
		return errors.Wrapf(err, "failed to find agent")
	}

	var messages []entity.Message
	if err := sess.
		Order("created_at DESC").
		Limit(200).
		Find(&messages, "thread_id = ?", threadId).Error; err != nil {
		return errors.Wrapf(err, "failed to find messages")
	}

	slices.SortStableFunc(messages, func(a, b entity.Message) int {
		if a.CreatedAt.Before(b.CreatedAt) {
			return -1
		} else if a.CreatedAt.After(b.CreatedAt) {
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
				Thread:              &thread,
			}

			// build recent conversations
			for _, msg := range messages {
				instValues.RecentConversations = append(instValues.RecentConversations, Conversation{
					User: msg.User,
					Text: msg.Content.Data().Text,
					Actions: gog.Map(msg.Content.Data().ToolCalls, func(tc entity.MessageContentToolCall) string {
						return tc.Name
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
				tools = append(tools, s.toolManager.GetLocalTool(ctx, tool.Name))
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
				myctx.WithThread(
					myctx.WithAgent(
						ctx,
						&agent,
					),
					&thread,
				),
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

			var conversation Conversation
			if err := json.Unmarshal([]byte(responseText), &conversation); err != nil {
				return errors.Wrapf(err, "failed to unmarshal conversation")
			}

			content := entity.MessageContent{
				Text: conversation.Text,
			}

			toolCallData := tool.GetCallData(ctx)
			for _, data := range toolCallData {
				tc := entity.MessageContentToolCall{
					Name:      data.Name,
					Arguments: data.Arguments,
					Result:    data.Result,
				}
				content.ToolCalls = append(content.ToolCalls, tc)
			}

			newMessage := entity.Message{
				ThreadID: threadId,
				User:     agent.Name,
				Content:  datatypes.NewJSONType(content),
			}
			if err := sess.Create(&newMessage).Error; err != nil {
				return errors.Wrapf(err, "failed to create message")
			}

			return nil
		})
	}

	return eg.Wait()
}
