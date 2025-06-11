package engine_test

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/engine"
)

func (s *EngineTestSuite) TestGenerate() {
	ctx := context.Background()

	var out string
	resp, err := s.engine.Generate(
		ctx,
		&engine.GenerateRequest{
			Model: "gpt-4o",
		},
		&out,
		ai.WithPrompt("Hello, world!"),
	)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	s.Require().Contains(out, "Hello")
}
