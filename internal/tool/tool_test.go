package tool_test

import (
	"testing"

	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/habiliai/agentruntime/internal/tool"
	"github.com/jcooky/go-din"
	"github.com/stretchr/testify/suite"
)

type ToolTestSuite struct {
	mytesting.Suite

	toolManager tool.Manager
}

func (s *ToolTestSuite) SetupTest() {
	s.Suite.SetupTest()

	s.toolManager = din.MustGetT[tool.Manager](s.Container)
}

func (s *ToolTestSuite) TearDownTest() {
	s.Suite.TearDownTest()
}

func TestTool(t *testing.T) {
	suite.Run(t, new(ToolTestSuite))
}
