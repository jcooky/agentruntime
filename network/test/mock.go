package networktest

import (
	"context"

	"github.com/habiliai/agentruntime/network"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AgentNetworkClientMock struct {
	mock.Mock
}

// CheckLive implements network.AgentNetworkClient.
func (m *AgentNetworkClientMock) CheckLive(ctx context.Context, in *network.CheckLiveRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// DeregisterAgent implements network.AgentNetworkClient.
func (m *AgentNetworkClientMock) DeregisterAgent(ctx context.Context, in *network.DeregisterAgentRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// RegisterAgent implements network.AgentNetworkClient.
func (m *AgentNetworkClientMock) RegisterAgent(ctx context.Context, in *network.RegisterAgentRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *AgentNetworkClientMock) GetAgentRuntimeInfo(ctx context.Context, in *network.GetAgentRuntimeInfoRequest, opts ...grpc.CallOption) (*network.GetAgentRuntimeInfoResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*network.GetAgentRuntimeInfoResponse), args.Error(1)
}

var (
	_ network.AgentNetworkClient = (*AgentNetworkClientMock)(nil)
)
