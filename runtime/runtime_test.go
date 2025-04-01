package runtime_test

import (
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/habiliai/agentruntime/thread"
	threadtest "github.com/habiliai/agentruntime/thread/test"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type AgentRuntimeTestSuite struct {
	mytesting.Suite

	agents        []config.AgentConfig
	runtime       runtime.Service
	threadManager *threadtest.ThreadManagerClientMock
}

func (s *AgentRuntimeTestSuite) SetupTest() {
	os.Setenv("ENV_TEST_FILE", "../.env.test")
	s.Suite.SetupTest()

	var err error

	s.agents, err = config.LoadAgentsFromFiles([]string{"./testdata/test1.agent.yaml"})
	s.Require().NoError(err)

	s.threadManager = &threadtest.ThreadManagerClientMock{}
	di.Set(s.Context, thread.ClientKey, s.threadManager)
	s.runtime = di.MustGet[runtime.Service](s, runtime.ServiceKey)

	s.Require().NoError(err)
}

func (s *AgentRuntimeTestSuite) TearDownTest() {
	defer s.Suite.TearDownTest()
}

func TestAgentRuntime(t *testing.T) {
	suite.Run(t, new(AgentRuntimeTestSuite))
}
