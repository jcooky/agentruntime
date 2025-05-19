package networktest

import (
	"context"

	"github.com/habiliai/agentruntime/network"
	"github.com/stretchr/testify/mock"
)

type JsonRpcClient struct {
	mock.Mock
}

func (t *JsonRpcClient) CreateThread(ctx context.Context, request *network.CreateThreadRequest) (*network.CreateThreadResponse, error) {
	args := t.Called(ctx, request)
	return args.Get(0).(*network.CreateThreadResponse), args.Error(1)
}

func (t *JsonRpcClient) GetThread(ctx context.Context, request *network.GetThreadRequest) (*network.Thread, error) {
	args := t.Called(ctx, request)
	return args.Get(0).(*network.Thread), args.Error(1)
}

func (t *JsonRpcClient) AddMessage(ctx context.Context, request *network.AddMessageRequest) (*network.AddMessageResponse, error) {
	args := t.Called(ctx, request)
	return args.Get(0).(*network.AddMessageResponse), args.Error(1)
}

func (t *JsonRpcClient) GetMessages(ctx context.Context, request *network.GetMessagesRequest) (*network.GetMessagesResponse, error) {
	args := t.Called(ctx, request)
	return args.Get(0).(*network.GetMessagesResponse), args.Error(1)
}

func (t *JsonRpcClient) GetNumMessages(ctx context.Context, request *network.GetNumMessagesRequest) (*network.GetNumMessagesResponse, error) {
	args := t.Called(ctx, request)
	return args.Get(0).(*network.GetNumMessagesResponse), args.Error(1)
}

func (j *JsonRpcClient) CheckLive(ctx context.Context, request *network.CheckLiveRequest) error {
	args := j.Called(ctx, request)
	return args.Error(0)
}

func (j *JsonRpcClient) GetAgentRuntimeInfo(ctx context.Context, request *network.GetAgentRuntimeInfoRequest) (*network.GetAgentRuntimeInfoResponse, error) {
	args := j.Called(ctx, request)
	return args.Get(0).(*network.GetAgentRuntimeInfoResponse), args.Error(1)
}

func (j *JsonRpcClient) RegisterAgent(ctx context.Context, request *network.RegisterAgentRequest) error {
	args := j.Called(ctx, request)
	return args.Error(0)
}

func (j *JsonRpcClient) DeregisterAgent(ctx context.Context, request *network.DeregisterAgentRequest) error {
	args := j.Called(ctx, request)
	return args.Error(0)
}

func (j *JsonRpcClient) IsMentionedOnce(ctx context.Context, request *network.IsMentionedRequest) (*network.IsMentionedResponse, error) {
	args := j.Called(ctx, request)
	return args.Get(0).(*network.IsMentionedResponse), args.Error(1)
}

var (
	_ network.JsonRpcClient = (*JsonRpcClient)(nil)
)
