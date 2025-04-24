package engine_test

import (
	_ "embed"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/stretchr/testify/suite"
	"os"
	"strings"
	"testing"
)

//go:embed testdata/test1.agent.yaml
var test1AgentYaml string

type EngineTestSuite struct {
	mytesting.Suite

	engine      engine.Engine
	agentConfig config.AgentConfig
}

func (s *EngineTestSuite) SetupTest() {
	os.Setenv("ENV_TEST_FILE", "../.env.test")
	s.Suite.SetupTest()

	var err error

	s.agentConfig, err = config.LoadAgentFromFile(strings.NewReader(test1AgentYaml))
	s.Require().NoError(err)

	s.engine = di.MustGet[engine.Engine](s.Context, engine.Key)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(EngineTestSuite))
}
