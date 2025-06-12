package cmd

import (
	"context"
	"log/slog"
	"strings"

	"github.com/habiliai/agentruntime"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/internal/msgutils"
	"github.com/mokiat/gog"
	"gorm.io/gorm"
)

func loopMentionedBy(
	ctx context.Context,
	db *gorm.DB,
	runtimes map[string]*agentruntime.AgentRuntime,
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
			if db.First(&msg.Thread, "id = ?", msg.ThreadID).Error != nil {
				logger.Error("thread not found", "thread_id", msg.ThreadID)
				continue
			}

			for _, mention := range mentions {
				runtime, ok := runtimes[strings.ToLower(mention)]
				if !ok {
					logger.Error("runtime not found", "mention", mention)
					continue
				}

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
