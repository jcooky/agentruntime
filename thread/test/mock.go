package threadtest

import (
	"context"
	grpcutils_testing "github.com/habiliai/agentruntime/internal/grpcutils/testing"
	"github.com/habiliai/agentruntime/thread"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type ThreadManagerClientMock struct {
	mock.Mock
}

func (t *ThreadManagerClientMock) CreateThread(ctx context.Context, in *thread.CreateThreadRequest, _ ...grpc.CallOption) (*thread.CreateThreadResponse, error) {
	args := t.Called(ctx, in)
	return args.Get(0).(*thread.CreateThreadResponse), args.Error(1)
}

func (t *ThreadManagerClientMock) GetThread(ctx context.Context, in *thread.GetThreadRequest, _ ...grpc.CallOption) (*thread.Thread, error) {
	args := t.Called(ctx, in)
	return args.Get(0).(*thread.Thread), args.Error(1)
}

func (t *ThreadManagerClientMock) AddMessage(ctx context.Context, in *thread.AddMessageRequest, _ ...grpc.CallOption) (*thread.AddMessageResponse, error) {
	args := t.Called(ctx, in)
	return args.Get(0).(*thread.AddMessageResponse), args.Error(1)
}

// MockGetMessagesClient is a mock of ThreadManager_GetMessagesClient
type MockGetMessagesClient struct {
	grpcutils_testing.ClientStreamMock
}

func (m *MockGetMessagesClient) Recv() (*thread.GetMessagesResponse, error) {
	args := m.Called()
	return args.Get(0).(*thread.GetMessagesResponse), args.Error(1)
}

func (t *ThreadManagerClientMock) GetMessages(ctx context.Context, in *thread.GetMessagesRequest, _ ...grpc.CallOption) (thread.ThreadManager_GetMessagesClient, error) {
	args := t.Called(ctx, in)
	return args.Get(0).(thread.ThreadManager_GetMessagesClient), args.Error(1)
}

func (t *ThreadManagerClientMock) GetNumMessages(ctx context.Context, in *thread.GetNumMessagesRequest, _ ...grpc.CallOption) (*thread.GetNumMessagesResponse, error) {
	args := t.Called(ctx, in)
	return args.Get(0).(*thread.GetNumMessagesResponse), args.Error(1)
}

var (
	_ thread.ThreadManagerClient             = (*ThreadManagerClientMock)(nil)
	_ thread.ThreadManager_GetMessagesClient = (*MockGetMessagesClient)(nil)
)
