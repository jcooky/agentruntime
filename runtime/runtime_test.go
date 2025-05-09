package runtime_test

import (
	"os"
	"testing"

	"github.com/habiliai/agentruntime/network"
	"github.com/jcooky/go-din"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/mytesting"
	networktest "github.com/habiliai/agentruntime/network/test"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/stretchr/testify/suite"
)

type AgentRuntimeTestSuite struct {
	mytesting.Suite

	agents       []config.AgentConfig
	runtime      runtime.Service
	agentNetwork *networktest.JsonRpcClient
}

func (s *AgentRuntimeTestSuite) SetupTest() {
	os.Setenv("ENV_TEST_FILE", "../.env.test")
	s.Suite.SetupTest()

	var err error

	s.agents, err = config.LoadAgentsFromFiles([]string{"./testdata/test1.agent.yaml"})
	s.Require().NoError(err)

	s.agentNetwork = &networktest.JsonRpcClient{}
	din.SetT[network.JsonRpcClient](s.Container, s.agentNetwork)
	s.runtime = din.MustGetT[runtime.Service](s.Container)

	s.Require().NoError(err)
}

func (s *AgentRuntimeTestSuite) TearDownTest() {
	s.Suite.TearDownTest()
}

func TestAgentRuntime(t *testing.T) {
	suite.Run(t, new(AgentRuntimeTestSuite))
}
