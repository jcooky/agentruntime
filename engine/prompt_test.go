package engine_test

import (
	"context"

	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
)

func (s *EngineTestSuite) TestBuildPromptValues() {
	agent := entity.Agent{
		Name:      "Alice",
		Role:      "weather forecaster",
		Prompt:    "You are a weather forecaster. You can ask her about the weather in any city.",
		ModelName: "openai/gpt-4o",
		ModelConfig: map[string]any{
			"temperature": 0.5,
		},
		MessageExamples: [][]entity.MessageExample{
			{
				{
					User: "USER",
					Text: "What is the weather in Tokyo?",
				},
				{
					User:    "Alice",
					Text:    "Today's weather in Tokyo is sunny with a temperature of 20°C.",
					Actions: []string{"get_weather"},
				},
			},
		},
	}

	history := []engine.Conversation{
		{
			User: "USER",
			Text: "What is the weather in Tokyo?",
		},
	}

	thread := engine.Thread{
		Instruction: "User ask about the weather in specific city.",
		Participants: []engine.Participant{
			{
				Name: "Alice",
				Role: "Weather forecaster",
			},
		},
	}

	runRequest := engine.RunRequest{
		History:           history,
		ThreadInstruction: thread.Instruction,
		Participant:       thread.Participants,
	}

	promptValues, err := s.engine.BuildPromptValues(context.Background(), agent, runRequest)
	s.Require().NoError(err)
	s.Require().NotNil(promptValues)

	s.T().Logf(">> PromptValues: %v\n", promptValues)

	promptFn := engine.GetPromptFn(promptValues)
	s.Require().NotNil(promptFn)

	prompt, err := promptFn(context.Background(), nil)
	s.Require().NoError(err)
	s.Require().NotEmpty(prompt)

	s.T().Logf(">> Prompt: %s\n", prompt)
}
