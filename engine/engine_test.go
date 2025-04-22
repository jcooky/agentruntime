package engine_test

import (
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/stretchr/testify/suite"
	"testing"
)

type EngineTestSuite struct {
	mytesting.Suite

	engine      engine.Engine
	agentConfig config.AgentConfig
}

func (s *EngineTestSuite) SetupTest() {
	s.Suite.SetupTest()

	var err error

	s.agentConfig, err = config.LoadAgentFromFile("./testdata/test1.agent.yaml")
	s.Require().NoError(err)

	s.engine = di.MustGet[engine.Engine](s.Context, engine.Key)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(EngineTestSuite))
}
