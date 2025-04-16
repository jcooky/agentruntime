package runner_test

import (
	"github.com/habiliai/agentruntime/runner"
	"github.com/mokiat/gog"
)

func (s *RunnerTestSuite) TestRun() {
	ag, err := s.runner.NewAgentFromConfig(s, s.agentConfig)
	s.Require().NoError(err)

	resp, err := s.runner.Run(s, runner.RunRequest{
		ThreadInstruction: "# Mission: AI agents dialogue with user",
		History: []runner.Conversation{
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
	toolCallNames := gog.Map(resp.ToolCalls, func(tc runner.RunResponseToolcall) string {
		return tc.Name
	})
	s.Require().Contains(toolCallNames, "done_agent")
	s.Require().Contains(toolCallNames, "get_weather")

}
