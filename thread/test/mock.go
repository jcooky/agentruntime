package threadtest

import (
	"context"

	"github.com/habiliai/agentruntime/thread"
	"github.com/stretchr/testify/mock"
)

type JsonRpcClient struct {
	mock.Mock
}

func (t *JsonRpcClient) CreateThread(ctx context.Context, request *thread.CreateThreadRequest) (*thread.CreateThreadResponse, error) {
	args := t.Called(ctx, request)
	return args.Get(0).(*thread.CreateThreadResponse), args.Error(1)
}

func (t *JsonRpcClient) GetThread(ctx context.Context, request *thread.GetThreadRequest) (*thread.Thread, error) {
	args := t.Called(ctx, request)
	return args.Get(0).(*thread.Thread), args.Error(1)
}

func (t *JsonRpcClient) AddMessage(ctx context.Context, request *thread.AddMessageRequest) (*thread.AddMessageResponse, error) {
	args := t.Called(ctx, request)
	return args.Get(0).(*thread.AddMessageResponse), args.Error(1)
}

func (t *JsonRpcClient) GetMessages(ctx context.Context, request *thread.GetMessagesRequest) (*thread.GetMessagesResponse, error) {
	args := t.Called(ctx, request)
	return args.Get(0).(*thread.GetMessagesResponse), args.Error(1)
}

func (t *JsonRpcClient) GetNumMessages(ctx context.Context, request *thread.GetNumMessagesRequest) (*thread.GetNumMessagesResponse, error) {
	args := t.Called(ctx, request)
	return args.Get(0).(*thread.GetNumMessagesResponse), args.Error(1)
}

var (
	_ thread.JsonRpcClient = (*JsonRpcClient)(nil)
)
