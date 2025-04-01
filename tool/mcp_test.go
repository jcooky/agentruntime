package tool_test

import (
	"github.com/habiliai/agentruntime/tool"
)

func (s *ToolTestSuite) TestToolByMCP() {
	s.Require().NoError(s.toolManager.RegisterMCPTool(s, tool.RegisterMCPToolRequest{
		ServerName: "filesystem",
		Command:    "npx",
		Args: []string{
			"-y", "@modelcontextprotocol/server-filesystem", ".",
		},
	}))

	tool := s.toolManager.GetTool(s, "filesystem/list_directory")
	s.NotNil(tool)

	s.Equal("filesystem_list_directory", tool.Definition().Name)
	s.T().Logf("tool definition: %v", tool.Definition())
}
