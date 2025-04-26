package engine_test

import (
	"context"
	"github.com/habiliai/agentruntime/engine"
)

func (s *EngineTestSuite) TestGenerate() {
	ctx := context.Background()

	var out string
	resp, err := s.engine.Generate(
		ctx,
		engine.GenerateRequest{
			PromptTmpl: "Hello, world!",
			Model:      "gpt-4o",
		},
		&out,
	)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	s.Require().Contains(out, "Hello")
}
