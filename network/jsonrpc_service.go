package network

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/rpc/v2"
	"github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/thread"
	"github.com/jcooky/go-din"

	"github.com/habiliai/agentruntime/entity"
)

type (
	JsonRpcService struct {
		service       Service
		threadManager thread.Manager
	}

	CheckLiveRequest struct {
		Names []string `json:"names"`
	}

	DeregisterAgentRequest struct {
		Names []string `json:"names"`
	}

	RegisterAgentRequest struct {
		Addr   string       `json:"addr"`
		Secure bool         `json:"secure,omitempty"`
		Info   []*AgentInfo `json:"info"`
	}

	GetAgentRuntimeInfoRequest struct {
		Names []string `json:"names,omitempty"`
		All   bool     `json:"all,omitempty"`
	}

	GetAgentRuntimeInfoResponse struct {
		AgentRuntimeInfo []*AgentRuntimeInfo `json:"agent_runtime_info,omitempty"`
	}

	AgentInfo struct {
		Name     string            `json:"name"`
		Role     string            `json:"role"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}

	AgentRuntimeInfo struct {
		Info   *AgentInfo `json:"info"`
		Addr   string     `json:"addr"`
		Secure bool       `json:"secure,omitempty"`
	}

	GetMessagesRequest struct {
		ThreadId uint32 `json:"thread_id"`
		Order    string `json:"order" jsonschema:"enum:latest,oldest"`
		Limit    uint32 `json:"limit"`
		Cursor   uint32 `json:"cursor"`
	}

	MessageToolCall struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
		Result    string `json:"result"`
	}

	Message struct {
		Id        uint32             `json:"id"`
		Content   string             `json:"content"`
		CreatedAt time.Time          `json:"created_at"`
		UpdatedAt time.Time          `json:"updated_at"`
		Sender    string             `json:"sender"`
		ToolCalls []*MessageToolCall `json:"tool_calls"`
	}

	GetMessagesResponse struct {
		Messages   []*Message `json:"messages"`
		NextCursor uint32     `json:"next_cursor"`
	}

	GetNumMessagesRequest struct {
		ThreadId uint32 `json:"thread_id"`
	}

	GetNumMessagesResponse struct {
		NumMessages uint32 `json:"num_messages"`
	}

	CreateThreadRequest struct {
		Instruction string            `json:"instruction"`
		Metadata    map[string]string `json:"metadata"`
	}

	CreateThreadResponse struct {
		ThreadId uint32 `json:"thread_id"`
	}

	GetThreadRequest struct {
		ThreadId uint32 `json:"thread_id"`
	}

	Thread struct {
		Id           uint32    `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Instruction  string    `json:"instruction"`
		Participants []string  `json:"participants"`
	}

	AddMessageRequest struct {
		ThreadId  uint32             `json:"thread_id"`
		Sender    string             `json:"sender"`
		Content   string             `json:"content"`
		ToolCalls []*MessageToolCall `json:"tool_calls"`
	}

	AddMessageResponse struct {
		MessageId uint32 `json:"message_id"`
	}
)

func (s *JsonRpcService) CheckLive(r *http.Request, args *CheckLiveRequest, _ *struct{}) error {
	return s.service.CheckLive(r.Context(), args.Names)
}

func (s *JsonRpcService) GetAgentRuntimeInfo(r *http.Request, args *GetAgentRuntimeInfoRequest, reply *GetAgentRuntimeInfoResponse) error {
	var (
		runtimeInfo []entity.AgentRuntime
		err         error
	)
	if args.All {
		runtimeInfo, err = s.service.GetAllAgentRuntimeInfo(r.Context())
	} else {
		runtimeInfo, err = s.service.GetAgentRuntimeInfo(r.Context(), args.Names)
	}
	if err != nil {
		return err
	}

	reply.AgentRuntimeInfo = make([]*AgentRuntimeInfo, 0, len(runtimeInfo))
	for _, agent := range runtimeInfo {
		reply.AgentRuntimeInfo = append(reply.AgentRuntimeInfo, &AgentRuntimeInfo{
			Addr:   agent.Addr,
			Secure: agent.Secure,
			Info: &AgentInfo{
				Name:     agent.Name,
				Role:     agent.Role,
				Metadata: agent.Metadata.Data(),
			},
		})
	}

	return nil
}

func (s *JsonRpcService) RegisterAgent(r *http.Request, args *RegisterAgentRequest, _ *struct{}) error {
	if err := s.service.RegisterAgent(r.Context(), args.Addr, args.Secure, args.Info); err != nil {
		return err
	}

	return nil
}

func (s *JsonRpcService) DeregisterAgent(r *http.Request, args *DeregisterAgentRequest, _ *struct{}) error {
	if err := s.service.DeregisterAgent(r.Context(), args.Names); err != nil {
		return err
	}

	return nil
}

func (s *JsonRpcService) GetMessages(r *http.Request, args *GetMessagesRequest, reply *GetMessagesResponse) error {
	cursor := uint(args.Cursor)
	order := "ASC"
	if args.Order == "latest" {
		order = "DESC"
	}

	for {
		messages, err := s.threadManager.GetMessages(r.Context(), uint(args.ThreadId), order, cursor, uint(args.Limit))
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			break
		}

		for _, msg := range messages {
			content := msg.Content.Data()
			res := Message{
				Id:        uint32(msg.ID),
				Content:   content.Text,
				CreatedAt: msg.CreatedAt,
				UpdatedAt: msg.UpdatedAt,
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
				res.ToolCalls = append(res.ToolCalls, &MessageToolCall{
					Name:      toolCall.Name,
					Arguments: string(args),
					Result:    string(result),
				})
			}
			reply.Messages = append(reply.Messages, &res)
			cursor = msg.ID
		}
	}
	reply.NextCursor = uint32(cursor)

	return nil
}

func (s *JsonRpcService) GetNumMessages(r *http.Request, args *GetNumMessagesRequest, reply *GetNumMessagesResponse) error {
	numMessages, err := s.threadManager.GetNumMessages(r.Context(), uint(args.ThreadId))
	if err != nil {
		return err
	}

	reply.NumMessages = uint32(numMessages)
	return nil
}

func (m *JsonRpcService) CreateThread(r *http.Request, args *CreateThreadRequest, reply *CreateThreadResponse) error {
	thr, err := m.threadManager.CreateThread(r.Context(), args.Instruction)
	if err != nil {
		return err
	}

	reply.ThreadId = uint32(thr.ID)

	return nil
}

func (s *JsonRpcService) GetThread(r *http.Request, args *GetThreadRequest, reply *Thread) error {
	thr, err := s.threadManager.GetThreadById(r.Context(), uint(args.ThreadId))
	if err != nil {
		return err
	}

	reply.Id = uint32(thr.ID)
	reply.Instruction = thr.Instruction
	reply.CreatedAt = thr.CreatedAt
	reply.UpdatedAt = thr.UpdatedAt

	return nil
}

func (s *JsonRpcService) AddMessage(r *http.Request, args *AddMessageRequest, reply *AddMessageResponse) error {
	content := entity.MessageContent{
		Text: args.Content,
	}

	for _, toolCall := range args.ToolCalls {
		var args any
		if err := json.Unmarshal([]byte(toolCall.Arguments), &args); err != nil {
			return errors.Wrapf(err, "failed to unmarshal tool call arguments")
		}
		var result any
		if err := json.Unmarshal([]byte(toolCall.Result), &result); err != nil {
			return errors.Wrapf(err, "failed to unmarshal tool call result")
		}
		content.ToolCalls = append(content.ToolCalls, entity.MessageContentToolCall{
			Name:      toolCall.Name,
			Arguments: args,
			Result:    result,
		})
	}

	msg, err := s.threadManager.AddMessage(r.Context(), uint(args.ThreadId), args.Sender, content)
	if err != nil {
		return err
	}

	reply.MessageId = uint32(msg.ID)

	return nil
}

var (
	servicePrefix = "habiliai.agentnetwork.v1"
)

func RegisterJsonRpcService(c *din.Container, server *rpc.Server) error {
	svc := &JsonRpcService{
		service: din.MustGetT[Service](c),
	}

	return errors.Wrapf(server.RegisterService(svc, servicePrefix), "failed to register jsonrpc service")
}
