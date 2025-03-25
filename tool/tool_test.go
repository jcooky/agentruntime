package tool_test

import (
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/habiliai/agentruntime/tool"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ToolTestSuite struct {
	mytesting.Suite

	toolManager tool.Manager
}

func (s *ToolTestSuite) SetupTest() {
	s.Suite.SetupTest()
	s.Context = di.WithContainer(s.Context, di.EnvTest)

	s.toolManager = di.MustGet[tool.Manager](s, tool.ManagerKey)
}

func (s *ToolTestSuite) TearDownTest() {
	defer s.Suite.TearDownTest()
}

func TestTool(t *testing.T) {
	suite.Run(t, new(ToolTestSuite))
}
