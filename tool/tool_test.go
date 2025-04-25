package tool_test

import (
	"testing"

	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/habiliai/agentruntime/tool"
	"github.com/stretchr/testify/suite"
)

type ToolTestSuite struct {
	mytesting.Suite

	toolManager tool.Manager
}

func (s *ToolTestSuite) SetupTest() {
	s.Suite.SetupTest()

	s.toolManager = di.MustGet[tool.Manager](s, s.Container, tool.ManagerKey)
}

func (s *ToolTestSuite) TearDownTest() {
	s.Suite.TearDownTest()
}

func TestTool(t *testing.T) {
	suite.Run(t, new(ToolTestSuite))
}
