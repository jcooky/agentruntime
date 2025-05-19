package networktest

import (
	"context"

	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/network"
	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (s *ServiceMock) GetAgentRuntimeInfo(ctx context.Context, agentNames []string) ([]entity.AgentRuntime, error) {
	args := s.Called(ctx, agentNames)
	return args.Get(0).([]entity.AgentRuntime), args.Error(1)
}

func (s *ServiceMock) GetAllAgentRuntimeInfo(ctx context.Context) ([]entity.AgentRuntime, error) {
	args := s.Called(ctx)
	return args.Get(0).([]entity.AgentRuntime), args.Error(1)
}

func (s *ServiceMock) RegisterAgent(ctx context.Context, addr string, agentInfo []*network.AgentInfo) error {
	args := s.Called(ctx, addr, agentInfo)
	return args.Error(0)
}

func (s *ServiceMock) DeregisterAgent(ctx context.Context, agentNames []string) error {
	args := s.Called(ctx, agentNames)
	return args.Error(0)
}

func (s *ServiceMock) CheckLive(ctx context.Context, agentNames []string) error {
	args := s.Called(ctx, agentNames)
	return args.Error(0)
}

var (
	_ network.Service = (*ServiceMock)(nil) // Ensure ServiceMock implements the Service interface
)
