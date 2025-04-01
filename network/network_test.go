package network_test

import (
	"context"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/network"
	"github.com/stretchr/testify/suite"
	"testing"
)

type NetworkTestSuite struct {
	suite.Suite
	context.Context

	manager network.Service
}

func (s *NetworkTestSuite) SetupTest() {
	s.Context = context.TODO()
	s.Context = di.WithContainer(s.Context, di.EnvTest)

	s.manager = di.MustGet[network.Service](s.Context, network.ManagerKey)
}

func TestAgents(t *testing.T) {
	suite.Run(t, new(NetworkTestSuite))
}
