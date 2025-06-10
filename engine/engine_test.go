package engine_test

import (
	_ "embed"
	"os"
	"testing"

	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/stretchr/testify/suite"
)

//go:embed testdata/test1.agent.yaml
var test1AgentYaml string

type EngineTestSuite struct {
	mytesting.Suite

	engine engine.Engine
}

func (s *EngineTestSuite) SetupTest() {
	os.Setenv("ENV_TEST_FILE", "../.env.test")
	s.Suite.SetupTest()

	s.engine = engine.NewEngine(
		mylog.NewLogger(),
		tool.NewManager(),
		genkit.NewGenkit(),
	)
}

func (s *EngineTestSuite) TearDownTest() {
	s.Suite.TearDownTest()
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(EngineTestSuite))
}
