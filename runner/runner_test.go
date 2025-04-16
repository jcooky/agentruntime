package runner_test

import (
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/habiliai/agentruntime/runner"
	"github.com/stretchr/testify/suite"
	"testing"
)

type RunnerTestSuite struct {
	mytesting.Suite

	runner      runner.Runner
	agentConfig config.AgentConfig
}

func (s *RunnerTestSuite) SetupTest() {
	s.Suite.SetupTest()

	var err error

	s.agentConfig, err = config.LoadAgentFromFile("./testdata/test1.agent.yaml")
	s.Require().NoError(err)

	s.runner = di.MustGet[runner.Runner](s.Context, runner.Key)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(RunnerTestSuite))
}
