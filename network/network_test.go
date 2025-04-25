package network_test

import (
	"context"
	"testing"

	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/network"
	"github.com/stretchr/testify/suite"
)

type NetworkTestSuite struct {
	suite.Suite
	context.Context

	manager   network.Service
	container *di.Container
}

func (s *NetworkTestSuite) SetupTest() {
	s.Context = context.TODO()
	s.container = di.NewContainer(di.EnvTest)

	s.manager = di.MustGet[network.Service](s.Context, s.container, network.ManagerKey)
}

func TestAgents(t *testing.T) {
	suite.Run(t, new(NetworkTestSuite))
}
