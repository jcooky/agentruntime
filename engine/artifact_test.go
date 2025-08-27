package engine_test

import (
	"context"

	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
)

func (s *EngineTestSuite) TestArtifactGenerationInstructions() {
	// Test that the chat template contains artifact generation instructions when enabled
	agent := entity.Agent{
		Name:               "TestAgent",
		Role:               "test agent",
		Prompt:             "You are a test agent that can create artifacts.",
		ModelName:          "openai/gpt-4o",
		ArtifactGeneration: true,
	}

	thread := &engine.Thread{
		Instruction:  "Test thread for artifact generation",
		Participants: []engine.Participant{},
	}

	runRequest := engine.RunRequest{
		History:           []engine.Conversation{},
		ThreadInstruction: thread.Instruction,
		Participant:       thread.Participants,
	}

	// Build prompt values
	promptValues, err := s.engine.BuildPromptValues(context.Background(), agent, runRequest)
	s.Require().NoError(err)
	s.Require().NotNil(promptValues)

	// Get the actual prompt
	promptFn := engine.GetPromptFn(promptValues)
	s.Require().NotNil(promptFn)

	prompt, err := promptFn(context.Background(), nil)
	s.Require().NoError(err)
	s.Require().NotEmpty(prompt)

	// Test that artifact instruction XML block is properly structured
	s.Assert().Contains(prompt, `<artifact_instruction required="true">`)
	s.Assert().Contains(prompt, "</artifact_instruction>")

	// Test that artifact generation section is present
	s.Assert().Contains(prompt, "# ARTIFACT GENERATION:")
	s.Assert().Contains(prompt, "When to Use Artifacts")
	s.Assert().Contains(prompt, "Why Use Artifacts")
	s.Assert().Contains(prompt, "How to Create Artifacts")

	// Test that HTML method is documented
	s.Assert().Contains(prompt, "Universal HTML Method")

	// Test HTML chart example
	s.Assert().Contains(prompt, `<habili:artifact type="html" title="Sales Performance Chart">`)
	s.Assert().Contains(prompt, `<script src="https://cdn.jsdelivr.net/npm/chart.js@4"></script>`)

	// Test HTML code method
	s.Assert().Contains(prompt, `<habili:artifact type="html"`)
	s.Assert().Contains(prompt, `<htmlCode>`)
	s.Assert().Contains(prompt, `<!DOCTYPE html>`)

	// Test that CDN components are mentioned
	s.Assert().Contains(prompt, "CDN")
	s.Assert().Contains(prompt, "https://cdn.tailwindcss.com")

	// Test that Tailwind CSS classes are mentioned
	s.Assert().Contains(prompt, "Tailwind CSS (Direct Styling - CDN Loaded)")
	s.Assert().Contains(prompt, "bg-blue-600")

	// Test examples are present
	s.Assert().Contains(prompt, "Data Entry Form:")
	s.Assert().Contains(prompt, "Simple Game/Calculator:")

	s.T().Logf("✅ Artifact generation instructions are properly included in the template")
}

func (s *EngineTestSuite) TestArtifactExampleValidity() {
	// Test that the artifact examples in the template are syntactically valid
	agent := entity.Agent{
		Name:               "ExampleAgent",
		Role:               "example tester",
		Prompt:             "Test artifact examples",
		ModelName:          "openai/gpt-4o",
		ArtifactGeneration: true,
	}

	runRequest := engine.RunRequest{
		History: []engine.Conversation{},
	}

	promptValues, err := s.engine.BuildPromptValues(context.Background(), agent, runRequest)
	s.Require().NoError(err)

	promptFn := engine.GetPromptFn(promptValues)
	prompt, err := promptFn(context.Background(), nil)
	s.Require().NoError(err)

	// Check that examples have proper XML structure
	examples := []struct {
		name    string
		pattern string
	}{
		{
			name:    "HTML Chart Example",
			pattern: `<habili:artifact type="html" title="Sales Performance Chart">`,
		},
		{
			name:    "HTML Code Example",
			pattern: `<habili:artifact type="html"`,
		},
		{
			name:    "Contact Form Example",
			pattern: `<habili:artifact type="html" title="Contact Form">`,
		},
		{
			name:    "Number Guessing Game Example",
			pattern: `<habili:artifact type="html" title="Number Guessing Game">`,
		},
	}

	for _, example := range examples {
		s.Assert().Contains(prompt, example.pattern,
			"Should contain valid %s", example.name)
	}

	// Check that HTML code examples have proper structure
	htmlCodeChecks := []string{
		"<!DOCTYPE html>",
		"document.getElementById('contactForm')",
		"let target = Math.floor(Math.random() * 100) + 1;",
		"</htmlCode>",
		"</habili:artifact>",
	}

	for _, check := range htmlCodeChecks {
		s.Assert().Contains(prompt, check,
			"Should contain valid HTML code structure: %s", check)
	}

	// Verify React data structures are properly formatted
	s.Assert().Contains(prompt, `labels: [`, "Should contain properly formatted React data structure")
	s.Assert().Contains(prompt, `backgroundColor: '#`, "Should contain proper color specification in React code")

	s.T().Logf("✅ All artifact examples have valid syntax")
}

func (s *EngineTestSuite) TestTemplateXMLStructure() {
	// Test that all major template sections are properly structured with XML tags and attributes
	agent := entity.Agent{
		Name:               "StructureAgent",
		Role:               "structure tester",
		Prompt:             "Test XML structure",
		ModelName:          "openai/gpt-4o",
		ArtifactGeneration: true,
	}

	thread := &engine.Thread{
		Instruction: "Test XML structure",
		Participants: []engine.Participant{
			{Name: "TestParticipant", Role: "tester", Description: "Test participant"},
		},
	}

	runRequest := engine.RunRequest{
		History: []engine.Conversation{
			{User: "USER", Text: "Test message"},
		},
		ThreadInstruction: thread.Instruction,
		Participant:       thread.Participants,
	}

	promptValues, err := s.engine.BuildPromptValues(context.Background(), agent, runRequest)
	s.Require().NoError(err)

	promptFn := engine.GetPromptFn(promptValues)
	prompt, err := promptFn(context.Background(), nil)
	s.Require().NoError(err)

	// Test thread section structure
	s.Assert().Contains(prompt, `<thread dynamic="true">`)
	s.Assert().Contains(prompt, `</thread>`)

	// Test participants structure
	s.Assert().Contains(prompt, `<participants count="1">`)
	s.Assert().Contains(prompt, `<participant>`)
	s.Assert().Contains(prompt, `</participants>`)

	// Test agent section structure
	s.Assert().Contains(prompt, `<agent name="StructureAgent" model="openai/gpt-4o">`)
	s.Assert().Contains(prompt, `</agent>`)

	// Test history section structure
	s.Assert().Contains(prompt, `<history dynamic="true" optional="true">`)
	s.Assert().Contains(prompt, `</history>`)

	// Test available_actions structure
	s.Assert().Contains(prompt, `<available_actions dynamic="true">`)
	s.Assert().Contains(prompt, `</available_actions>`)

	// Test behavior_rules structure
	s.Assert().Contains(prompt, `<behavior_rules required="true">`)
	s.Assert().Contains(prompt, `</behavior_rules>`)

	// Test artifact_instruction structure
	s.Assert().Contains(prompt, `<artifact_instruction required="true">`)
	s.Assert().Contains(prompt, `</artifact_instruction>`)

	s.T().Logf("✅ All template sections are properly structured with XML tags and attributes")
}

func (s *EngineTestSuite) TestArtifactInstructionConditionalRendering() {
	s.Run("Agent with artifactGeneration enabled", func() {
		// Test that the artifact instruction is included when artifactGeneration is enabled
		agent := entity.Agent{
			Name:               "TestAgent",
			Description:        "Test agent for artifact generation",
			Role:               "Assistant",
			Prompt:             "You are a helpful assistant",
			ModelName:          "openai/gpt-4o",
			ArtifactGeneration: true,
		}

		runRequest := engine.RunRequest{
			History:           []engine.Conversation{},
			ThreadInstruction: "Test instruction",
			Participant:       []engine.Participant{},
		}

		promptValues, err := s.engine.BuildPromptValues(context.Background(), agent, runRequest)
		s.Require().NoError(err)

		promptFn := engine.GetPromptFn(promptValues)
		result, err := promptFn(context.Background(), nil)
		s.Require().NoError(err)

		// Check that artifact instruction section exists
		s.Assert().Contains(result, "# ARTIFACT GENERATION:")
		s.Assert().Contains(result, "## When to Use Artifacts")
		s.Assert().Contains(result, "Universal HTML Method")
		s.Assert().Contains(result, "<habili:artifact")
		s.Assert().Contains(result, `<artifact_instruction required="true">`)

		// Check that approved technologies are mentioned
		s.Assert().Contains(result, "chart.js")
		s.Assert().Contains(result, "Pure HTML")
		s.Assert().Contains(result, "Vanilla JavaScript")
		s.Assert().Contains(result, "Tailwind CSS")

		// Check security policy
		s.Assert().Contains(result, "ONLY the 3 approved technologies")
	})

	s.Run("Agent without artifactGeneration", func() {
		// Test that the artifact instruction is NOT included when artifactGeneration is not set
		agent := entity.Agent{
			Name:        "TestAgent",
			Description: "Test agent without artifact generation",
			Role:        "Assistant",
			Prompt:      "You are a helpful assistant",
			ModelName:   "openai/gpt-4o",
			Metadata:    map[string]any{},
		}

		runRequest := engine.RunRequest{
			History:           []engine.Conversation{},
			ThreadInstruction: "Test instruction",
			Participant:       []engine.Participant{},
		}

		promptValues, err := s.engine.BuildPromptValues(context.Background(), agent, runRequest)
		s.Require().NoError(err)

		promptFn := engine.GetPromptFn(promptValues)
		result, err := promptFn(context.Background(), nil)
		s.Require().NoError(err)

		// Check that artifact instruction section does NOT exist
		s.Assert().NotContains(result, "# ARTIFACT GENERATION:")
		s.Assert().NotContains(result, "## When to Use Artifacts")
		s.Assert().NotContains(result, "<habili:artifact")
		s.Assert().NotContains(result, `<artifact_instruction required="true">`)
	})

	s.Run("Agent with artifactGeneration set to false", func() {
		// Test that the artifact instruction is NOT included when artifactGeneration is explicitly false
		agent := entity.Agent{
			Name:        "TestAgent",
			Description: "Test agent with artifact generation disabled",
			Role:        "Assistant",
			Prompt:      "You are a helpful assistant",
			ModelName:   "openai/gpt-4o",
			Metadata: map[string]any{
				"artifactGeneration": false,
			},
		}

		runRequest := engine.RunRequest{
			History:           []engine.Conversation{},
			ThreadInstruction: "Test instruction",
			Participant:       []engine.Participant{},
		}

		promptValues, err := s.engine.BuildPromptValues(context.Background(), agent, runRequest)
		s.Require().NoError(err)

		promptFn := engine.GetPromptFn(promptValues)
		result, err := promptFn(context.Background(), nil)
		s.Require().NoError(err)

		// Check that artifact instruction section does NOT exist
		s.Assert().NotContains(result, "# ARTIFACT GENERATION:")
		s.Assert().NotContains(result, "## When to Use Artifacts")
		s.Assert().NotContains(result, "<habili:artifact")
		s.Assert().NotContains(result, `<artifact_instruction required="true">`)
	})

	s.T().Logf("✅ Artifact instruction conditional rendering works correctly")
}
