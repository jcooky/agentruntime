package tool_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/genkit"
	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/habiliai/agentruntime/internal/tool"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	mytesting.Suite

	toolManager tool.Manager
}

func (s *TestSuite) SetupTest() {
	s.Suite.SetupTest()

	g, err := genkit.NewGenkit(
		s,
		nil,
		slog.Default(),
		false,
	)
	s.Require().NoError(err)
	s.toolManager, err = tool.NewToolManager(
		s,
		[]entity.AgentSkill{
			{
				Type:    "mcp",
				Name:    "filesystem",
				Command: "npx",
				Args: []string{
					"-y", "@modelcontextprotocol/server-filesystem", ".",
				},
			},
			{
				Type:        "nativeTool",
				Name:        "get_weather",
				Description: "Get weather information when you need it",
				Env: map[string]string{
					"OPENWEATHER_API_KEY": os.Getenv("OPENWEATHER_API_KEY"),
				},
			},
		},
		slog.Default(),
		g,
	)
	s.Require().NoError(err)

}

func (s *TestSuite) TearDownTest() {
	s.toolManager.Close()
	s.Suite.TearDownTest()
}

func TestTool(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
