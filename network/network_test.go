package network_test

import (
	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/jcooky/go-din"
	"testing"

	"github.com/habiliai/agentruntime/network"
	"github.com/stretchr/testify/suite"
)

type NetworkTestSuite struct {
	mytesting.Suite

	manager network.Service
}

func (s *NetworkTestSuite) SetupTest() {
	s.Suite.SetupTest()

	s.manager = din.MustGetT[network.Service](s.Container)
}

func (s *NetworkTestSuite) TearDownTest() {
	s.Suite.TearDownTest()
}

func TestAgents(t *testing.T) {
	suite.Run(t, new(NetworkTestSuite))
}
