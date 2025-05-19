package runtimetest

import (
	"context"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/stretchr/testify/mock"
)

type RuntimeServiceMock struct {
	mock.Mock
}

func (m *RuntimeServiceMock) FindAgentsByNames(names []string) ([]entity.Agent, error) {
	args := m.Called(names)
	return args.Get(0).([]entity.Agent), args.Error(1)
}

func (m *RuntimeServiceMock) RegisterAgent(ctx context.Context, ac config.AgentConfig) (*entity.Agent, error) {
	args := m.Called(ctx, ac)
	return args.Get(0).(*entity.Agent), args.Error(1)
}

func (m *RuntimeServiceMock) Run(ctx context.Context, threadIds uint, agents []entity.Agent) error {
	args := m.Called(ctx, threadIds, agents)
	return args.Error(0)
}

var (
	_ runtime.Service = (*RuntimeServiceMock)(nil)
)
