package engine_test

import (
	"github.com/habiliai/agentruntime/engine"
	"github.com/mokiat/gog"
)

func (s *EngineTestSuite) TestRun() {
	ag, err := s.engine.NewAgentFromConfig(s, s.agentConfig)
	s.Require().NoError(err)

	resp, err := s.engine.Run(s, engine.RunRequest{
		ThreadInstruction: "# Mission: AI agents dialogue with user",
		History: []engine.Conversation{
			{
				User: "USER",
				Text: "Hello, what is the weather today?",
			},
		},
		Agent: *ag,
	})
	s.Require().NoError(err)
	s.T().Logf(">> RunResponse: %v\n", resp)

	if !s.Len(resp.ToolCalls, 2) {
		s.T().FailNow()
	}
	toolCallNames := gog.Map(resp.ToolCalls, func(tc engine.RunResponseToolcall) string {
		return tc.Name
	})
	s.Require().Contains(toolCallNames, "done_agent")
	s.Require().Contains(toolCallNames, "get_weather")

}
