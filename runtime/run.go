package runtime

import (
	"context"
	_ "embed"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/thread"
	"github.com/mokiat/gog"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"io"
	"slices"
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

			for _, msg := range resp.Messages {
				if msg.Sender == "USER" {
					continue
				}
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

	agentRuntimeInfo, err := s.networkClient.GetAgentRuntimeInfo(ctx, &network.GetAgentRuntimeInfoRequest{
		All: gog.PtrOf(true),
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
			Actions: gog.Map(msg.ToolCalls, func(tc *thread.Message_ToolCall) engine.Action {
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

			resp, err := s.runner.Run(ctx, engine.RunRequest{
				ThreadInstruction: thr.Instruction,
				History:           history,
				Agent:             agent,
				Participant:       participants,
			})
			if err != nil {
				return err
			}

			req := &thread.AddMessageRequest{
				ThreadId: uint32(threadId),
				Sender:   agent.Name,
				Content:  resp.Content,
			}

			for _, toolCall := range resp.ToolCalls {
				tc := thread.Message_ToolCall{
					Name:      toolCall.Name,
					Arguments: string(toolCall.Arguments),
					Result:    string(toolCall.Result),
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
