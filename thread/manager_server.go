package thread

import (
	"context"
	"encoding/json"

	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type managerServer struct {
	UnsafeThreadManagerServer

	manager Manager
}

func (m *managerServer) GetMessages(req *GetMessagesRequest, stream ThreadManager_GetMessagesServer) error {
	ctx := stream.Context()
	cursor := uint(0)
	order := "ASC"
	if req.Order == GetMessagesRequest_LATEST {
		order = "DESC"
	}
	for {
		messages, err := m.manager.GetMessages(ctx, uint(req.ThreadId), order, cursor, uint(req.Limit))
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			break
		}

		var resp GetMessagesResponse
		for _, msg := range messages {
			content := msg.Content.Data()
			res := Message{
				Id:        uint32(msg.ID),
				Content:   content.Text,
				CreatedAt: timestamppb.New(msg.CreatedAt),
				UpdatedAt: timestamppb.New(msg.UpdatedAt),
				Sender:    msg.User,
			}
			for _, toolCall := range content.ToolCalls {
				args, err := json.Marshal(toolCall.Arguments)
				if err != nil {
					return errors.Wrapf(err, "failed to marshal tool call arguments")
				}
				result, err := json.Marshal(toolCall.Result)
				if err != nil {
					return errors.Wrapf(err, "failed to marshal tool call result")
				}
				res.ToolCalls = append(res.ToolCalls, &Message_ToolCall{
					Name:      toolCall.Name,
					Arguments: string(args),
					Result:    string(result),
				})
			}
			resp.Messages = append(resp.Messages, &res)
		}

		if err := stream.Send(&resp); err != nil {
			return errors.Wrapf(err, "failed to send messages")
		}
		cursor = messages[len(messages)-1].ID
	}

	return nil
}

func (m *managerServer) GetNumMessages(ctx context.Context, req *GetNumMessagesRequest) (*GetNumMessagesResponse, error) {
	numMessages, err := m.manager.GetNumMessages(ctx, uint(req.ThreadId))
	if err != nil {
		return nil, err
	}

	return &GetNumMessagesResponse{
		NumMessages: uint32(numMessages),
	}, nil
}

func (m *managerServer) CreateThread(ctx context.Context, req *CreateThreadRequest) (*CreateThreadResponse, error) {
	thr, err := m.manager.CreateThread(ctx, req.Instruction)
	if err != nil {
		return nil, err
	}

	return &CreateThreadResponse{
		ThreadId: uint32(thr.ID),
	}, nil
}

func (m *managerServer) GetThread(ctx context.Context, request *GetThreadRequest) (*Thread, error) {
	thr, err := m.manager.GetThreadById(ctx, uint(request.ThreadId))
	if err != nil {
		return nil, err
	}

	return &Thread{
		Id:          uint32(thr.ID),
		Instruction: thr.Instruction,
		CreatedAt:   timestamppb.New(thr.CreatedAt),
		UpdatedAt:   timestamppb.New(thr.UpdatedAt),
	}, nil
}

func (m *managerServer) AddMessage(ctx context.Context, request *AddMessageRequest) (*AddMessageResponse, error) {
	content := entity.MessageContent{
		Text: request.Content,
	}

	for _, toolCall := range request.ToolCalls {
		var args any
		if err := json.Unmarshal([]byte(toolCall.Arguments), &args); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal tool call arguments")
		}
		var result any
		if err := json.Unmarshal([]byte(toolCall.Result), &result); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal tool call result")
		}
		content.ToolCalls = append(content.ToolCalls, entity.MessageContentToolCall{
			Name:      toolCall.Name,
			Arguments: args,
			Result:    result,
		})
	}

	msg, err := m.manager.AddMessage(ctx, uint(request.ThreadId), request.Sender, content)
	if err != nil {
		return nil, err
	}

	return &AddMessageResponse{
		MessageId: uint32(msg.ID),
	}, nil
}

var (
	_                ThreadManagerServer = (*managerServer)(nil)
	ManagerServerKey                     = di.NewKey()
)

func init() {
	di.Register(ManagerServerKey, func(ctx context.Context, container *di.Container) (any, error) {
		return &managerServer{
			manager: di.MustGet[Manager](ctx, container, ManagerKey),
		}, nil
	})
}
