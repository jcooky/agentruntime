package tool_test

func (s *TestSuite) TestToolByMCP() {
	tool := s.toolManager.GetMCPTool("filesystem", "list_directory")
	s.NotNil(tool)

	s.Equal("list_directory", tool.Definition().Name)
	s.T().Logf("tool definition: %v", tool.Definition())
}
