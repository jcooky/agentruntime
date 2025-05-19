package runtime

import (
	"context"
	_ "embed"
	"slices"

	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/network"
	"github.com/mokiat/gog"
	"golang.org/x/sync/errgroup"
)

func (s *service) getMessages(
	ctx context.Context,
	threadId uint,
) (res []*network.Message, err error) {
	reply, err := s.networkClient.GetMessages(ctx, &network.GetMessagesRequest{
		ThreadId: uint32(threadId),
		Order:    "latest",
		Limit:    200,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get messages")
	}

	return reply.Messages, nil
}

func (s *service) Run(
	ctx context.Context,
	threadId uint,
	agents []entity.Agent,
) error {
	thr, err := s.networkClient.GetThread(ctx, &network.GetThreadRequest{
		ThreadId: uint32(threadId),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to get thread")
	}

	messages, err := s.getMessages(ctx, threadId)
	if err != nil {
		return err
	}

	slices.SortStableFunc(messages, func(a, b *network.Message) int {
		if a.CreatedAt.Before(b.CreatedAt) {
			return -1
		} else if a.CreatedAt.After(b.CreatedAt) {
			return 1
		} else {
			return 0
		}
	})

	agentRuntimeInfo, err := s.networkClient.GetAgentRuntimeInfo(ctx, &network.GetAgentRuntimeInfoRequest{
		All: true,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to get agent runtime info")
	}

	agentInfoMap := make(map[string]*network.AgentInfo)
	for _, agentRuntime := range agentRuntimeInfo.AgentRuntimeInfo {
		agentInfoMap[agentRuntime.Info.Name] = agentRuntime.Info
	}

	// build recent conversations
	var (
		history      = make([]engine.Conversation, 0, len(messages))
		participants []engine.Participant
	)
	for _, msg := range messages {
		history = append(history, engine.Conversation{
			User: msg.Sender,
			Text: msg.Content,
			Actions: gog.Map(msg.ToolCalls, func(tc *network.MessageToolCall) engine.Action {
				return engine.Action{
					Name:      tc.Name,
					Arguments: tc.Arguments,
					Result:    tc.Result,
				}
			}),
		})
		if sender, ok := agentInfoMap[msg.Sender]; ok {
			participants = append(participants, engine.Participant{
				Name: sender.Name,
				Role: sender.Role,
			})
		}
	}

	var eg errgroup.Group
	for _, agent := range agents {
		eg.Go(func() error {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			var content string
			resp, err := s.runner.Run(ctx, engine.RunRequest{
				ThreadInstruction: thr.Instruction,
				History:           history,
				Agent:             agent,
				Participant:       participants,
			}, &content)
			if err != nil {
				return err
			}

			req := &network.AddMessageRequest{
				ThreadId: uint32(threadId),
				Sender:   agent.Name,
				Content:  content,
			}

			for _, toolCall := range resp.ToolCalls {
				tc := network.MessageToolCall{
					Name:      toolCall.Name,
					Arguments: string(toolCall.Arguments),
					Result:    string(toolCall.Result),
				}

				req.ToolCalls = append(req.ToolCalls, &tc)
			}

			// add message to thread
			if _, err := s.networkClient.AddMessage(ctx, req); err != nil {
				return errors.Wrapf(err, "failed to add message")
			}

			return nil
		})
	}

	return eg.Wait()
}
