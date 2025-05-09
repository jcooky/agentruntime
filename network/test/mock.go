package networktest

import (
	"context"

	"github.com/habiliai/agentruntime/network"
	"github.com/stretchr/testify/mock"
)

type JsonRpcClient struct {
	mock.Mock
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

var (
	_ network.JsonRpcClient = (*JsonRpcClient)(nil)
)
