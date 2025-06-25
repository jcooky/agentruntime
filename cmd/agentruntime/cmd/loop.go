package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/habiliai/agentruntime"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/msgutils"
	"github.com/mokiat/gog"
	"gorm.io/gorm"
)

func loopMentionedBy(
	ctx context.Context,
	db *gorm.DB,
	agents map[string]entity.Agent,
	logger *slog.Logger,
	messageCh chan *Message,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-messageCh:
			mentions := msgutils.ExtractMentions(msg.Content)
			if len(mentions) == 0 {
				continue
			}
			if db.Preload("History").First(&msg.Thread, "id = ?", msg.ThreadID).Error != nil {
				logger.Error("thread not found", "thread_id", msg.ThreadID)
				continue
			}

			for _, mention := range mentions {
				agent, ok := agents[mention]
				if !ok {
					logger.Error("agent not found", "mention", mention)
					continue
				}
				runtime, err := agentruntime.NewAgentRuntime(
					ctx,
					agentruntime.WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
					agentruntime.WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
					agentruntime.WithXAIAPIKey(os.Getenv("XAI_API_KEY")),
					agentruntime.WithLogger(logger),
					agentruntime.WithTraceVerbose(true),
					agentruntime.WithAgent(agent),
				)
				if err != nil {
					logger.Error("failed to create agent runtime", "mention", mention, "error", err)
					continue
				}
				defer runtime.Close()

				var out string
				resp, err := runtime.Run(ctx, engine.RunRequest{
					ThreadInstruction: msg.Thread.Instruction,
					History: gog.Map(msg.Thread.History, func(m Message) engine.Conversation {
						return engine.Conversation{
							User: m.User,
							Text: m.Content,
							Actions: gog.Map(m.Actions, func(a Action) engine.Action {
								return engine.Action{
									Name:      a.Name,
									Arguments: a.Args,
									Result:    a.Result,
								}
							}),
						}
					}),
					Participant: gog.Map(msg.Thread.Participants, func(p string) engine.Participant {
						return engine.Participant{
							Name: p,
						}
					}),
				}, &out)
				if err != nil {
					logger.Error("failed to run agent", "mention", mention, "error", err)
					continue
				}

				actions := gog.Map(resp.ToolCalls, func(t engine.ToolCall) Action {
					return Action{
						Name:   t.Name,
						Args:   t.Arguments,
						Result: t.Result,
					}
				})

				msg := &Message{
					ThreadID: msg.ThreadID,
					Content:  out,
					User:     runtime.Agent().Name,
					Actions:  actions,
				}
				if err := db.Create(msg).Error; err != nil {
					logger.Error("failed to create message", "error", err)
					continue
				}

				messageCh <- msg
			}
		}
	}
}
