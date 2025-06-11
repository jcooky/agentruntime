package engine_test

import (
	_ "embed"
	"log/slog"
	"os"
	"testing"

	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/internal/genkit"
	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/habiliai/agentruntime/internal/tool"
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

	g, err := genkit.NewGenkit(s, s.Config(), slog.Default(), true)
	s.Require().NoError(err)

	toolManager, err := tool.NewToolManager(s.Context(), s.Config(), slog.Default(), g)
	s.Require().NoError(err)

	s.engine = engine.NewEngine(
		slog.Default(),
		toolManager,
		genkit.NewGenkit(),
	)
}

func (s *EngineTestSuite) TearDownTest() {
	s.Suite.TearDownTest()
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(EngineTestSuite))
}
