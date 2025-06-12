package engine_test

import (
	_ "embed"
	"log/slog"
	"os"
	"testing"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/internal/genkit"
	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/habiliai/agentruntime/memory"
	"github.com/stretchr/testify/suite"
)

type EngineTestSuite struct {
	mytesting.Suite

	engine *engine.Engine
}

func (s *EngineTestSuite) SetupTest() {
	s.Suite.SetupTest()

	g, err := genkit.NewGenkit(s, &config.ModelConfig{
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		XAIAPIKey:       os.Getenv("XAI_API_KEY"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}, slog.Default(), true)
	s.Require().NoError(err)

	memoryService, err := memory.NewService(s, &config.MemoryConfig{
		SqliteEnabled: true,
		SqlitePath:    ":memory:",
		VectorEnabled: true,
	}, slog.Default(), g)
	s.Require().NoError(err)

	s.engine = engine.NewEngine(
		slog.Default(),
		nil,
		g,
		memoryService,
	)
}

func (s *EngineTestSuite) TearDownTest() {
	s.Suite.TearDownTest()
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(EngineTestSuite))
}
