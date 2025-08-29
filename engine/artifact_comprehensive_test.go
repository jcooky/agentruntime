package engine_test

import (
	"strings"
	"testing"

	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
)

func (s *EngineTestSuite) TestHtmlCodeDirectGeneration() {
	// Test that the template supports direct HTML code generation
	agent := entity.Agent{
		Name:               "HtmlAgent",
		Role:               "html developer",
		Prompt:             "Create HTML components for users",
		ModelName:          "openai/gpt-4o",
		ArtifactGeneration: true,
	}

	runRequest := engine.RunRequest{
		History: []engine.Conversation{
			{
				User: "USER",
				Text: "Create an interactive counter component",
			},
		},
	}

	promptValues, err := s.engine.BuildPromptValues(s.T().Context(), agent, runRequest, nil)
	s.Require().NoError(err)

	promptFn := engine.GetPromptFn(promptValues)
	prompt, err := promptFn(s.T().Context(), nil)
	s.Require().NoError(err)

	// Test that html type is supported for all artifacts (charts, tables, components)
	s.Assert().Contains(prompt, `type="html"`, "Should support html artifacts")

	// Test HTML code specific instructions
	htmlInstructions := []string{
		"<htmlCode>",
		"</htmlCode>",
		"<!DOCTYPE html>",
		"addEventListener",
		"document.getElementById",
		"Universal HTML Method",
	}

	for _, instruction := range htmlInstructions {
		s.Assert().Contains(prompt, instruction,
			"Should contain HTML instruction: %s", instruction)
	}

	s.T().Logf("✅ HTML code direct generation is properly supported")
}

func (s *EngineTestSuite) TestXMLAttributesWithoutVersions() {
	// Test that XML attributes work correctly without version numbers
	agent := entity.Agent{
		Name:               "XMLAgent",
		Role:               "xml structure tester",
		Prompt:             "Test XML structure",
		ModelName:          "openai/gpt-4o",
		ArtifactGeneration: true,
	}

	thread := &engine.Thread{
		Instruction: "Test XML attributes",
		Participants: []engine.Participant{
			{Name: "TestUser", Role: "user", Description: "Test participant"},
		},
	}

	runRequest := engine.RunRequest{
		History: []engine.Conversation{
			{User: "USER", Text: "Test XML structure"},
		},
		ThreadInstruction: thread.Instruction,
		Participant:       thread.Participants,
	}

	promptValues, err := s.engine.BuildPromptValues(s.T().Context(), agent, runRequest, nil)
	s.Require().NoError(err)

	promptFn := engine.GetPromptFn(promptValues)
	prompt, err := promptFn(s.T().Context(), nil)
	s.Require().NoError(err)

	// Test that XML tags have meaningful attributes but no version numbers
	xmlChecks := map[string]string{
		"<thread dynamic=\"true\">":                         "Thread should have dynamic attribute",
		"<participants count=\"1\">":                        "Participants should have count attribute",
		"<agent name=\"XMLAgent\" model=\"openai/gpt-4o\">": "Agent should have name and model attributes",
		"<behavior_rules required=\"true\">":                "Behavior rules should have required attribute",
		"<artifact_instruction required=\"true\">":          "Artifact instruction should have required attribute",
	}

	for xmlTag, description := range xmlChecks {
		s.Assert().Contains(prompt, xmlTag, description)
	}

	// Ensure no version attributes exist
	versionChecks := []string{
		`version="1.0"`,
		`version="2.0"`,
		`version=`,
	}

	for _, versionAttr := range versionChecks {
		s.Assert().NotContains(prompt, versionAttr,
			"Should not contain version attribute: %s", versionAttr)
	}

	s.T().Logf("✅ XML attributes work correctly without version numbers")
}

func (s *EngineTestSuite) TestChartDesignGuidelineScenarios() {
	// Test various chart design guideline scenarios
	agent := entity.Agent{
		Name:               "ChartDesignAgent",
		Role:               "chart designer",
		Prompt:             "You create beautiful charts following design guidelines. Use appropriate color palettes for different scenarios.",
		ModelName:          "openai/gpt-4o",
		ArtifactGeneration: true,
	}

	scenarios := []struct {
		name             string
		userRequest      string
		expectedPalette  string
		expectedFeatures []string
	}{
		{
			name:            "Multi-Series Chart",
			userRequest:     "Create a chart showing revenue for Apple, Google, and Microsoft from 2020-2024",
			expectedPalette: "basicColorPalette",
			expectedFeatures: []string{
				"diverse items/categories",
				"#4F6D7A", // Slate Blue
				"#3C9D9B", // Muted Teal
				"#E4B363", // Golden Mustard
			},
		},
		{
			name:            "Professional Monochrome",
			userRequest:     "Show quarterly data in black and white professional style",
			expectedPalette: "greyModePalette",
			expectedFeatures: []string{
				"monochrome/professional",
				"#333333", // Deep Slate
				"#666666", // Charcoal Grey
				"#999999", // Neutral Grey
			},
		},
		{
			name:            "Emphasis Single Item",
			userRequest:     "Create a chart showing OpenAI revenue with emphasis, compared to other companies",
			expectedPalette: "primaryEmphasisPalette",
			expectedFeatures: []string{
				"highlighting single item",
				"#007ACC", // Primary Accent Blue
				"#777777", // Supporting grey
				"#BBBBBB", // Lighter supporting grey
			},
		},
	}

	for _, scenario := range scenarios {
		s.T().Run(scenario.name, func(t *testing.T) {
			runRequest := engine.RunRequest{
				History: []engine.Conversation{
					{User: "USER", Text: scenario.userRequest},
				},
			}

			promptValues, err := s.engine.BuildPromptValues(s.T().Context(), agent, runRequest, nil)
			s.Require().NoError(err)

			promptFn := engine.GetPromptFn(promptValues)
			prompt, err := promptFn(s.T().Context(), nil)
			s.Require().NoError(err)

			// Verify the scenario instructions contain the expected palette
			s.Assert().Contains(prompt, scenario.expectedPalette,
				"Should contain palette: %s for scenario: %s", scenario.expectedPalette, scenario.name)

			// Verify expected features are documented
			for _, feature := range scenario.expectedFeatures {
				s.Assert().Contains(prompt, feature,
					"Should contain feature: %s for scenario: %s", feature, scenario.name)
			}

			s.T().Logf("✅ Chart design scenario '%s' has proper guidance", scenario.name)
		})
	}
}

func (s *EngineTestSuite) TestChartJSConfigurationStandards() {
	// Test that Chart.js configuration standards are properly documented
	agent := entity.Agent{
		Name:               "ChartConfigAgent",
		Role:               "chart configuration specialist",
		Prompt:             "You configure Chart.js with proper design standards",
		ModelName:          "openai/gpt-4o",
		ArtifactGeneration: true,
	}

	runRequest := engine.RunRequest{
		History: []engine.Conversation{
			{User: "USER", Text: "Show me how to configure a professional chart"},
		},
	}

	promptValues, err := s.engine.BuildPromptValues(s.T().Context(), agent, runRequest, nil)
	s.Require().NoError(err)

	promptFn := engine.GetPromptFn(promptValues)
	prompt, err := promptFn(s.T().Context(), nil)
	s.Require().NoError(err)

	// Test Chart.js design standards
	configChecks := map[string][]string{
		"Line Chart Standards": {
			"borderWidth: 1",
			"tension: 0",
			"fill: false",
			"pointRadius: 3",
		},
		"Mass Chart Standards": {
			"borderWidth: 1",
			"radius: 0",
			"same color for border and fill",
		},
		"Layout Standards": {
			"backgroundColor: '#FAFAFA'",
			"plotAreaColor: '#FFFFFF'",
			"axisLineColor: '#999999'",
			"gridLineColor: '#DDDDDD'",
		},
		"Grid Configuration": {
			"color: '#DDDDDD'",
			"lineWidth: 1",
			"opacity: 0.6",
		},
		"Axis Configuration": {
			"bottom: true",
			"left: true",
			"top: false",
			"right: false",
		},
	}

	for section, checks := range configChecks {
		for _, check := range checks {
			s.Assert().Contains(prompt, check,
				"Should contain %s standard: %s", section, check)
		}
	}

	// Verify updated Chart.js examples use new standards
	s.Assert().Contains(prompt, "basicColorPalette", "Examples should use basic color palette")
	s.Assert().Contains(prompt, "grid: { color: '#DDDDDD', lineWidth: 1 }", "Examples should use proper grid config")
	s.Assert().Contains(prompt, "border: { color: '#999999', width: 1 }", "Examples should use proper border config")

	s.T().Logf("✅ Chart.js configuration standards are properly documented")
}

func (s *EngineTestSuite) TestTemplateRenderingVariousScenarios() {
	// Test template rendering in different scenarios
	scenarios := []struct {
		name        string
		agent       entity.Agent
		hasThread   bool
		hasHistory  bool
		hasExamples bool
	}{
		{
			name: "Minimal Agent",
			agent: entity.Agent{
				Name:               "MinimalAgent",
				Role:               "minimal",
				Prompt:             "Simple agent",
				ModelName:          "openai/gpt-4o",
				ArtifactGeneration: true,
			},
			hasThread:   false,
			hasHistory:  false,
			hasExamples: false,
		},
		{
			name: "Full Featured Agent",
			agent: entity.Agent{
				Name:               "FullAgent",
				Role:               "full featured",
				Prompt:             "Complex agent with examples",
				ModelName:          "anthropic/claude-4-sonnet",
				Description:        "A comprehensive agent for testing",
				ArtifactGeneration: true,
			},
			hasThread:   true,
			hasHistory:  true,
			hasExamples: true,
		},
	}

	for _, scenario := range scenarios {
		s.T().Run(scenario.name, func(t *testing.T) {
			var runRequest engine.RunRequest

			if scenario.hasThread {
				thread := &engine.Thread{
					Instruction: "Complex scenario test",
					Participants: []engine.Participant{
						{Name: "User1", Role: "user", Description: "Test user"},
						{Name: "User2", Role: "admin", Description: "Admin user"},
					},
				}
				runRequest.ThreadInstruction = thread.Instruction
				runRequest.Participant = thread.Participants
			}

			if scenario.hasHistory {
				runRequest.History = []engine.Conversation{
					{User: "USER", Text: "Previous message 1"},
					{User: "AGENT", Text: "Previous response 1"},
					{User: "USER", Text: "Previous message 2"},
				}
			}

			promptValues, err := s.engine.BuildPromptValues(s.T().Context(), scenario.agent, runRequest, nil)
			s.Require().NoError(err)

			promptFn := engine.GetPromptFn(promptValues)
			prompt, err := promptFn(s.T().Context(), nil)
			s.Require().NoError(err)

			// Basic structure should always be present
			s.Assert().Contains(prompt, "# ARTIFACT GENERATION:", "Should always contain artifact generation section")
			s.Assert().Contains(prompt, "<artifact_instruction required=\"true\">", "Should always contain artifact instruction")
			s.Assert().Contains(prompt, "<behavior_rules required=\"true\">", "Should always contain behavior rules")

			// Thread-specific checks
			if scenario.hasThread {
				s.Assert().Contains(prompt, "<thread dynamic=\"true\">", "Should contain thread section when thread exists")
				s.Assert().Contains(prompt, "<participants count=\"2\">", "Should contain correct participant count")
			} else {
				// For minimal agents, thread section should exist but be minimal
				s.Assert().Contains(prompt, "<thread dynamic=\"true\">", "Thread section should always exist")
				s.Assert().NotContains(prompt, "<participants count=\"2\">", "Should not contain complex participant structure")
			}

			// History-specific checks
			if scenario.hasHistory {
				s.Assert().Contains(prompt, "<history dynamic=\"true\" optional=\"true\">", "Should contain history section when history exists")
				s.Assert().Contains(prompt, "Previous message", "Should contain actual history content")
			}

			s.T().Logf("✅ Scenario '%s' rendered correctly", scenario.name)
		})
	}
}

func (s *EngineTestSuite) TestArtifactInstructionCompleteness() {
	// Comprehensive test for artifact instruction completeness
	agent := entity.Agent{
		Name:               "CompletenessAgent",
		Role:               "completeness tester",
		Prompt:             "Test completeness",
		ModelName:          "anthropic/claude-4-sonnet",
		ArtifactGeneration: true,
	}

	runRequest := engine.RunRequest{
		History: []engine.Conversation{
			{User: "USER", Text: "Create artifacts for me"},
		},
	}

	promptValues, err := s.engine.BuildPromptValues(s.T().Context(), agent, runRequest, nil)
	s.Require().NoError(err)

	promptFn := engine.GetPromptFn(promptValues)
	prompt, err := promptFn(s.T().Context(), nil)
	s.Require().NoError(err)

	// Extract artifact section
	artifactStart := strings.Index(prompt, "# ARTIFACT GENERATION:")
	s.Require().Greater(artifactStart, -1, "Should contain artifact generation section")

	artifactEnd := strings.Index(prompt[artifactStart:], "</artifact_instruction>")
	s.Require().Greater(artifactEnd, -1, "Should have proper artifact section ending")

	artifactSection := prompt[artifactStart : artifactStart+artifactEnd]

	// Test section structure
	requiredSections := []string{
		"## When to Use Artifacts",
		"## Why Use Artifacts",
		"## How to Create Artifacts",
		"### Universal HTML Method",
		"## Available Components & Styling",
		"## Best Practices",
		"## Examples by Use Case",
		"## Natural Language Integration",
	}

	for _, section := range requiredSections {
		s.Assert().Contains(artifactSection, section,
			"Should contain required section: %s", section)
	}

	// Test specific technical content
	technicalContent := []string{
		"CDN",
		"Tailwind CSS (Direct Styling - CDN Loaded)",
		"Chart.js",
		"Always Use Complete HTML Document",
		"Use Vanilla JavaScript for State Management",
		"Apply Responsive Design",
		"Handle User Interactions",
	}

	for _, content := range technicalContent {
		s.Assert().Contains(artifactSection, content,
			"Should contain technical content: %s", content)
	}

	// Test example completeness
	exampleChecks := []string{
		"Data Entry Form",
		"Number Guessing Game",
		"Interactive Dashboard",
		"Contact Form",
	}

	for _, example := range exampleChecks {
		s.Assert().Contains(artifactSection, example,
			"Should contain example: %s", example)
	}

	// Verify artifact section is comprehensive (length check)
	s.Assert().Greater(len(artifactSection), 25000,
		"Artifact section should be comprehensive (>25000 chars with new chart guidelines), got %d", len(artifactSection))

	// Verify comprehensive chart design guidelines
	s.Assert().Contains(artifactSection, "=== CHART COLOR PALETTES ===", "Should contain chart color palettes section")
	s.Assert().Contains(artifactSection, "=== CHART.JS DESIGN SETTINGS ===", "Should contain Chart.js design settings section")
	s.Assert().Contains(artifactSection, "=== COLOR SELECTION LOGIC ===", "Should contain color selection logic section")

	// Verify all three color palettes are defined
	basicPaletteColors := []string{"#4F6D7A", "#3C9D9B", "#E4B363", "#E17A72", "#6DA34D", "#7A5C92", "#5B7C99", "#C46BAE"}
	for _, color := range basicPaletteColors {
		s.Assert().Contains(artifactSection, color, "Should contain basic palette color: %s", color)
	}

	greyPaletteColors := []string{"#333333", "#4D4D4D", "#666666", "#808080", "#999999", "#B3B3B3", "#CCCCCC", "#E5E5E5"}
	for _, color := range greyPaletteColors {
		s.Assert().Contains(artifactSection, color, "Should contain grey palette color: %s", color)
	}

	primaryColors := []string{"#007ACC", "#777777", "#999999", "#BBBBBB", "#DDDDDD", "#F0F0F0"}
	for _, color := range primaryColors {
		s.Assert().Contains(artifactSection, color, "Should contain primary emphasis color: %s", color)
	}

	// Verify chart design configuration objects
	s.Assert().Contains(artifactSection, "chartLayout", "Should contain chart layout configuration")
	s.Assert().Contains(artifactSection, "lineChartDefaults", "Should contain line chart defaults")
	s.Assert().Contains(artifactSection, "massChartDefaults", "Should contain mass chart defaults")
	s.Assert().Contains(artifactSection, "gridConfig", "Should contain grid configuration")
	s.Assert().Contains(artifactSection, "axisConfig", "Should contain axis configuration")

	// Verify color selection function
	s.Assert().Contains(artifactSection, "getChartColors", "Should contain color selection function")

	s.T().Logf("✅ Artifact instructions are complete and comprehensive (%d chars)", len(artifactSection))
}

func (s *EngineTestSuite) TestArtifactGenerationEndToEnd() {
	// End-to-end test simulating actual artifact generation
	agent := entity.Agent{
		Name:               "E2EAgent",
		Role:               "end-to-end tester",
		Prompt:             "You are an AI that creates interactive artifacts for users. Create helpful visualizations and components.",
		ModelName:          "anthropic/claude-4-sonnet",
		ArtifactGeneration: true,
	}

	scenarios := []struct {
		name     string
		userText string
		expected []string
	}{
		{
			name:     "Chart Request",
			userText: "Create a bar chart showing monthly sales data",
			expected: []string{"chart", "data visualization", "monthly"},
		},
		{
			name:     "Interactive Component",
			userText: "I need a calculator component for my app",
			expected: []string{"react", "interactive", "component"},
		},
		{
			name:     "Data Table",
			userText: "Show this data in a table format",
			expected: []string{"table", "data", "format"},
		},
	}

	for _, scenario := range scenarios {
		s.T().Run(scenario.name, func(t *testing.T) {
			runRequest := engine.RunRequest{
				History: []engine.Conversation{
					{User: "USER", Text: scenario.userText},
				},
			}

			promptValues, err := s.engine.BuildPromptValues(s.T().Context(), agent, runRequest, nil)
			s.Require().NoError(err)

			promptFn := engine.GetPromptFn(promptValues)
			prompt, err := promptFn(s.T().Context(), nil)
			s.Require().NoError(err)

			// Verify the prompt contains all necessary instructions for handling this type of request
			s.Assert().Contains(prompt, "# ARTIFACT GENERATION:", "Should provide artifact generation instructions")

			// Check that the prompt provides guidance for the specific scenario
			for _, expectedContent := range scenario.expected {
				s.Assert().Contains(strings.ToLower(prompt), strings.ToLower(expectedContent),
					"Should contain guidance for: %s", expectedContent)
			}

			s.T().Logf("✅ End-to-end scenario '%s' has proper instruction coverage", scenario.name)
		})
	}
}
