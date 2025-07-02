package agentruntime_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/habiliai/agentruntime"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/knowledge"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/require"
)

func TestAgentRuntime(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test because OPENAI_API_KEY is not set")
	}
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping test because ANTHROPIC_API_KEY is not set")
	}

	bytes, err := os.ReadFile("examples/filesystem.agent.yaml")
	require.NoError(t, err)

	var agent entity.Agent
	err = yaml.Unmarshal(bytes, &agent)
	require.NoError(t, err)

	// Debug: Print the agent struct to see what fields are populated
	t.Logf("Agent struct: %+v", agent)
	t.Logf("Agent.Name: '%s'", agent.Name)
	t.Logf("Agent.Description: '%s'", agent.Description)
	t.Logf("Agent.ModelName: '%s'", agent.ModelName)
	t.Logf("Agent.System: '%s'", agent.System)
	t.Logf("Agent.Role: '%s'", agent.Role)
	t.Logf("Agent.Prompt: '%s'", agent.Prompt)
	t.Logf("Agent.Skills: %+v", agent.Skills)

	// Parse YAML as map to see the raw structure
	var yamlMap map[string]interface{}
	err = yaml.Unmarshal(bytes, &yamlMap)
	require.NoError(t, err)
	t.Logf("YAML map: %+v", yamlMap)

	// Test basic agent information
	require.Equal(t, "Bob", agent.Name, "Agent name should be 'Bob'")
	require.Contains(t, agent.Description, "Bob is a filesystem assistant", "Agent description should contain filesystem assistant info")
	require.Equal(t, "openai/gpt-4o", agent.ModelName, "Model name should be 'openai/gpt-4o'")
	require.Equal(t, "Take a deep breath and relax. Think step by step.", agent.System, "System prompt should match")
	require.Equal(t, "Assistant for Filesystem", agent.Role, "Role should be 'Assistant for Filesystem'")

	// Test prompt contains key instructions
	require.Contains(t, agent.Prompt, "Your name is Bob", "Prompt should contain name instruction")
	require.Contains(t, agent.Prompt, "control the file system", "Prompt should contain file system control instruction")
	require.Contains(t, agent.Prompt, "read, write, create, delete", "Prompt should contain file operations")

	// Test message examples
	require.Len(t, agent.MessageExamples, 2, "Should have 2 message examples")

	// Test first message example (read file)
	firstExample := agent.MessageExamples[0]
	require.Len(t, firstExample, 2, "First example should have 2 messages (user and agent)")
	require.Contains(t, firstExample[0].Text, "config.json", "First user message should mention config.json")
	require.Contains(t, firstExample[1].Text, "read the config.json", "First agent response should mention reading config.json")
	require.Contains(t, firstExample[1].Actions, "read_file", "First agent response should have read_file action")

	// Test second message example (write file)
	secondExample := agent.MessageExamples[1]
	require.Len(t, secondExample, 2, "Second example should have 2 messages (user and agent)")
	require.Contains(t, secondExample[0].Text, "hello.txt", "Second user message should mention hello.txt")
	require.Contains(t, secondExample[0].Text, "Hello World", "Second user message should mention Hello World content")
	require.Contains(t, secondExample[1].Text, "create the hello.txt", "Second agent response should mention creating hello.txt")
	require.Contains(t, secondExample[1].Actions, "write_file", "Second agent response should have write_file action")

	// Test skills
	require.Len(t, agent.Skills, 1, "Should have 1 skill")
	skill := agent.Skills[0]
	require.Equal(t, "mcp", skill.Type, "Skill type should be 'mcp'")
	require.Equal(t, "filesystem", skill.Name, "Skill name should be 'filesystem'")
	require.Equal(t, "npx", skill.Command, "Skill command should be 'npx'")
	require.Len(t, skill.Args, 3, "Skill should have 3 arguments")
	require.Equal(t, "-y", skill.Args[0], "First arg should be '-y'")
	require.Equal(t, "@modelcontextprotocol/server-filesystem", skill.Args[1], "Second arg should be MCP filesystem server")
	require.Equal(t, "./", skill.Args[2], "Third arg should be './'")

	agent.Skills = []entity.AgentSkill{}
	agent.ModelName = "anthropic/claude-4-sonnet"

	runtime, err := agentruntime.NewAgentRuntime(
		context.TODO(),
		agentruntime.WithAgent(agent),
		agentruntime.WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
		agentruntime.WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
		agentruntime.WithLogger(slog.Default()),
	)
	require.NoError(t, err)
	defer runtime.Close()

	var out string
	resp, err := runtime.Run(context.TODO(), engine.RunRequest{
		ThreadInstruction: "User ask about the weather in specific city.",
		History: []engine.Conversation{
			{
				User: "USER",
				Text: "Find and read the content of README.md file.",
			},
		},
	}, &out)
	require.NoError(t, err)
	require.NotNil(t, resp)

	t.Logf("Response: %+v", resp)
}

func TestAgentRuntimeWithLLMSkill(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test because OPENAI_API_KEY is not set")
	}

	bytes, err := os.ReadFile("examples/llm_agent.yaml")
	require.NoError(t, err)

	var agent entity.Agent
	err = yaml.Unmarshal(bytes, &agent)
	require.NoError(t, err)

	// Test basic agent information
	require.Equal(t, "Lily", agent.Name, "Agent name should be 'Lily'")
	require.Contains(t, agent.Description, "creative writing assistant", "Agent description should contain creative writing assistant info")
	require.Equal(t, "openai/gpt-4o", agent.ModelName, "Model name should be 'openai/gpt-4o'")
	require.Contains(t, agent.System, "creative writing assistant", "System prompt should mention creative writing")
	require.Equal(t, "Creative Writing Assistant", agent.Role, "Role should be 'Creative Writing Assistant'")

	// Test prompt contains key instructions
	require.Contains(t, agent.Prompt, "Your name is Lily", "Prompt should contain name instruction")
	require.Contains(t, agent.Prompt, "creative writing", "Prompt should contain creative writing instruction")
	require.Contains(t, agent.Prompt, "stories, poems", "Prompt should contain content types")

	// Test message examples
	require.Len(t, agent.MessageExamples, 2, "Should have 2 message examples")

	// Test first message example (story writing)
	firstExample := agent.MessageExamples[0]
	require.Len(t, firstExample, 2, "First example should have 2 messages (user and agent)")
	require.Contains(t, firstExample[0].Text, "robot learning to paint", "First user message should mention robot story")
	require.Contains(t, firstExample[1].Text, "creative writing expertise", "First agent response should mention creative writing")
	require.Contains(t, firstExample[1].Actions, "creative_writing_helper", "First agent response should have creative_writing_helper action")

	// Test second message example (poetry)
	secondExample := agent.MessageExamples[1]
	require.Len(t, secondExample, 2, "Second example should have 2 messages (user and agent)")
	require.Contains(t, secondExample[0].Text, "haiku about the ocean", "Second user message should mention haiku")
	require.Contains(t, secondExample[1].Text, "poetry skills", "Second agent response should mention poetry skills")
	require.Contains(t, secondExample[1].Actions, "poetry_generator", "Second agent response should have poetry_generator action")

	// Test LLM skills
	require.Len(t, agent.Skills, 2, "Should have 2 skills")

	// Test first LLM skill (creative_writing_helper)
	firstSkill := agent.Skills[0]
	require.Equal(t, "llm", firstSkill.Type, "First skill type should be 'llm'")
	require.Equal(t, "creative_writing_helper", firstSkill.Name, "First skill name should be 'creative_writing_helper'")
	require.Contains(t, firstSkill.Description, "creative writing assistance", "First skill description should mention creative writing assistance")
	require.Contains(t, firstSkill.Instruction, "narrative structure", "First skill instruction should mention narrative structure")
	require.Contains(t, firstSkill.Instruction, "character development", "First skill instruction should mention character development")

	// Test second LLM skill (poetry_generator)
	secondSkill := agent.Skills[1]
	require.Equal(t, "llm", secondSkill.Type, "Second skill type should be 'llm'")
	require.Equal(t, "poetry_generator", secondSkill.Name, "Second skill name should be 'poetry_generator'")
	require.Contains(t, secondSkill.Description, "poetry", "Second skill description should mention poetry")
	require.Contains(t, secondSkill.Instruction, "haiku, sonnet", "Second skill instruction should mention poetry forms")
	require.Contains(t, secondSkill.Instruction, "rhythm", "Second skill instruction should mention rhythm")

	// Test runtime creation and execution
	runtime, err := agentruntime.NewAgentRuntime(
		context.TODO(),
		agentruntime.WithAgent(agent),
		agentruntime.WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
		agentruntime.WithLogger(slog.Default()),
	)
	require.NoError(t, err)
	defer runtime.Close()

	var out string
	resp, err := runtime.Run(context.TODO(), engine.RunRequest{
		ThreadInstruction: "User asks for creative writing help.",
		History: []engine.Conversation{
			{
				User: "USER",
				Text: "Help me create a haiku about coding. Make it funny and relatable for programmers.",
			},
		},
	}, &out)
	require.NoError(t, err)
	require.NotNil(t, resp)

	t.Logf("Response: %+v", resp)
	t.Logf("Output: %s", out)

	// Verify the agent used the poetry_generator tool
	require.NotEmpty(t, resp.ToolCalls, "Should have at least one tool call")

	// Check if poetry_generator was called
	poetryGeneratorCalled := false
	for _, toolCall := range resp.ToolCalls {
		if toolCall.Name == "poetry_generator" {
			poetryGeneratorCalled = true

			// Verify the tool call result contains the instruction
			var result map[string]interface{}
			err = json.Unmarshal(toolCall.Result, &result)
			require.NoError(t, err, "Should be able to unmarshal tool result")

			instruction, ok := result["additional_important_instruction"].(string)
			require.True(t, ok, "Result should have additional_important_instruction field")
			require.Contains(t, instruction, "haiku", "Instruction should mention haiku")
			require.Contains(t, instruction, "rhythm", "Instruction should mention rhythm")

			t.Logf("Tool call arguments: %s", string(toolCall.Arguments))
			t.Logf("Tool call result: %s", string(toolCall.Result))
		}
	}

	require.True(t, poetryGeneratorCalled, "poetry_generator tool should have been called")

	// Verify the output contains haiku
	require.Contains(t, out, "haiku", "Output should mention haiku")
	require.NotEmpty(t, out, "Output should not be empty")
}

func TestAgentRuntimeWithEx1(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test because OPENAI_API_KEY is not set")
	}
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping test because ANTHROPIC_API_KEY is not set")
	}

	// Read JSON file instead of YAML
	bytes, err := os.ReadFile("examples/ex1.agent.json")
	require.NoError(t, err)

	var agent entity.Agent
	err = json.Unmarshal(bytes, &agent)
	require.NoError(t, err)

	// Test basic agent information
	require.Equal(t, "Edan", agent.Name, "Agent name should be 'Edan'")
	require.Contains(t, agent.Description, "summarizes startup deal info", "Agent description should contain startup summary info")
	require.Equal(t, "anthropic/claude-4-sonnet", agent.ModelName, "Model name should be 'anthropic/claude-4-sonnet'")
	require.Equal(t, "Moderator", agent.Role, "Role should be 'Moderator'")

	// Test prompt contains key instructions
	require.Contains(t, agent.Prompt, "moderator who specializes in summarizing startup deal information", "Prompt should contain moderator instruction")
	require.Contains(t, agent.Prompt, "team, traction, market, and competition", "Prompt should contain key analysis areas")

	// Test message examples
	require.Len(t, agent.MessageExamples, 1, "Should have 1 message example")
	firstExample := agent.MessageExamples[0]
	require.Len(t, firstExample, 2, "Example should have 2 messages (user and agent)")
	require.Contains(t, firstExample[0].Text, "summarize this startup's key points", "User message should ask for startup summary")
	require.Contains(t, firstExample[1].Actions, "startup-summary", "Agent response should have startup-summary action")

	// Test LLM skills
	require.Len(t, agent.Skills, 1, "Should have 1 skill")
	skill := agent.Skills[0]
	require.Equal(t, "llm", skill.Type, "Skill type should be 'llm'")
	require.Equal(t, "startup-summary", skill.Name, "Skill name should be 'startup-summary'")
	require.Contains(t, skill.Description, "Summarizes startup deal info", "Skill description should mention startup summary")
	require.Contains(t, skill.Instruction, "team, traction, market analysis", "Skill instruction should mention key analysis areas")

	// Test runtime creation and execution
	// Debug: Print API key availability
	t.Logf("ANTHROPIC_API_KEY set: %v", os.Getenv("ANTHROPIC_API_KEY") != "")
	t.Logf("OPENAI_API_KEY set: %v", os.Getenv("OPENAI_API_KEY") != "")

	runtime, err := agentruntime.NewAgentRuntime(
		context.TODO(),
		agentruntime.WithAgent(agent),
		agentruntime.WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
		agentruntime.WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
		agentruntime.WithLogger(slog.Default()),
	)
	require.NoError(t, err)
	defer runtime.Close()

	var out string
	resp, err := runtime.Run(context.TODO(), engine.RunRequest{
		ThreadInstruction: "User asks for startup analysis and summary.",
		History: []engine.Conversation{
			{
				User: "USER",
				Text: `I have a startup pitch that I need help analyzing:
				
				Company: TechFlow AI
				Team: 3 ex-Google engineers with 10+ years experience in ML
				Product: AI-powered code review tool that automatically suggests improvements
				Traction: 500 beta users, $10K MRR, growing 20% month-over-month
				Market: Developer tools market worth $20B, growing at 15% annually
				Competition: GitHub Copilot, Amazon CodeWhisperer
				Funding: Seeking $2M seed round
				
				Can you summarize this for investment review?`,
			},
		},
	}, &out)
	require.NoError(t, err)
	require.NotNil(t, resp)

	t.Logf("Response: %+v", resp)
	t.Logf("Output: %s", out)

	// Verify the agent used the startup-summary tool
	require.NotEmpty(t, resp.ToolCalls, "Should have at least one tool call")

	// Check if startup-summary was called
	startupSummaryCalled := false
	for _, toolCall := range resp.ToolCalls {
		if toolCall.Name == "startup-summary" {
			startupSummaryCalled = true

			// Verify the tool call result contains the instruction
			var result map[string]interface{}
			err = json.Unmarshal(toolCall.Result, &result)
			require.NoError(t, err, "Should be able to unmarshal tool result")

			instruction, ok := result["additional_important_instruction"].(string)
			require.True(t, ok, "Result should have additional_important_instruction field")
			require.Contains(t, instruction, "team, traction, market analysis", "Instruction should mention key analysis areas")

			t.Logf("Tool call arguments: %s", string(toolCall.Arguments))
			t.Logf("Tool call result: %s", string(toolCall.Result))
		}
	}

	require.True(t, startupSummaryCalled, "startup-summary tool should have been called")

	// Verify the output contains startup analysis elements
	require.Contains(t, out, "TechFlow AI", "Output should mention the company name")
	require.NotEmpty(t, out, "Output should not be empty")
}

func TestAgentRuntimeWithDennis(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test because OPENAI_API_KEY is not set")
	}
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping test because ANTHROPIC_API_KEY is not set")
	}

	bytes, err := os.ReadFile("examples/dennis.agent.json")
	require.NoError(t, err)

	var agent entity.Agent
	err = json.Unmarshal(bytes, &agent)
	require.NoError(t, err)

	runtime, err := agentruntime.NewAgentRuntime(
		context.TODO(),
		agentruntime.WithAgent(agent),
		agentruntime.WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
		agentruntime.WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
		agentruntime.WithLogger(slog.Default()),
	)
	require.NoError(t, err)
	defer runtime.Close()

	var out string
	resp, err := runtime.Run(context.TODO(), engine.RunRequest{
		ThreadInstruction: "User asks for startup analysis and summary.",
		History: []engine.Conversation{
			{
				User: "USER",
				Text: "Help me create a haiku about coding. Make it funny and relatable for programmers.",
			},
		},
	}, &out)

	require.NoError(t, err)
	require.NotNil(t, resp)

	t.Logf("Response: %+v", resp)
	t.Logf("Output: %s", out)
}

func TestAgentWithKnowledgeService(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test because OPENAI_API_KEY is not set")
	}
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping test because ANTHROPIC_API_KEY is not set")
	}

	bytes, err := os.ReadFile("examples/example_knowledge.agent.yaml")
	require.NoError(t, err)

	var agent entity.Agent
	err = yaml.Unmarshal(bytes, &agent)
	require.NoError(t, err)

	// Test basic agent information
	require.Equal(t, "HosuAgent", agent.Name, "Agent name should be 'HosuAgent'")
	require.Contains(t, agent.Description, "rescue dogs", "Agent description should contain rescue dogs info")
	require.Equal(t, "anthropic/claude-4-sonnet", agent.ModelName, "Model name should be 'anthropic/claude-4-sonnet'")
	require.Equal(t, "Rescue Dog Knowledge Assistant", agent.Role, "Role should be 'Rescue Dog Knowledge Assistant'")

	// Test system and prompt
	require.Contains(t, agent.System, "rescue dogs and animal welfare", "System should mention rescue dogs")
	require.Contains(t, agent.Prompt, "HosuAgent", "Prompt should contain agent name")
	require.Contains(t, agent.Prompt, "pet adoption", "Prompt should mention pet adoption")

	// Test knowledge entries
	require.Len(t, agent.Knowledge, 4, "Should have 4 knowledge entries")

	// Test first knowledge entry (Hosu info)
	hosuKnowledge := agent.Knowledge[0]
	require.Equal(t, "Hosu", hosuKnowledge["dogName"], "First knowledge should be about Hosu")
	require.Equal(t, "Mandu, Hoshu, Hodol", hosuKnowledge["aliases"], "Hosu aliases should match")
	require.Equal(t, "Mixed breed", hosuKnowledge["breed"], "Hosu breed should be Mixed breed")
	require.Equal(t, "3 years", hosuKnowledge["age"], "Hosu age should be 3 years")

	// Test second knowledge entry (Nuri shelter)
	shelterKnowledge := agent.Knowledge[1]
	require.Equal(t, "Nuri", shelterKnowledge["hometown"], "Second knowledge should be about Nuri shelter")
	require.Equal(t, "Seoul, South Korea", shelterKnowledge["location"], "Shelter location should be Seoul")
	require.Equal(t, "50 dogs", shelterKnowledge["capacity"], "Shelter capacity should be 50 dogs")

	// Test message examples
	require.Len(t, agent.MessageExamples, 2, "Should have 2 message examples")
	firstExample := agent.MessageExamples[0]
	require.Contains(t, firstExample[0].Text, "Tell me about Hosu", "First example should ask about Hosu")
	// First example should have knowledge_search action since knowledge is accessed via tool
	require.Contains(t, firstExample[1].Actions, "knowledge_search", "First example should use knowledge_search")

	// Second example should have web_search action
	secondExample := agent.MessageExamples[1]
	require.Contains(t, secondExample[0].Text, "adopting a rescue dog", "Second example should ask about adoption")
	require.Contains(t, secondExample[1].Actions, "web_search", "Second example should use web_search")

	// Test skills - now should include knowledge_search skill
	require.GreaterOrEqual(t, len(agent.Skills), 3, "Should have at least 3 skills")

	// Find and test the skills
	var webSearchSkill *entity.AgentSkill
	var adoptionAdvisorSkill *entity.AgentSkill
	var knowledgeSearchSkill *entity.AgentSkill
	for i, skill := range agent.Skills {
		switch skill.Name {
		case "web_search":
			webSearchSkill = &agent.Skills[i]
		case "adoption_advisor":
			adoptionAdvisorSkill = &agent.Skills[i]
		case "knowledge_search":
			knowledgeSearchSkill = &agent.Skills[i]
		}
	}
	require.NotNil(t, webSearchSkill, "Should have web_search skill")
	require.Equal(t, "nativeTool", webSearchSkill.Type, "web_search should be nativeTool type")
	require.Contains(t, webSearchSkill.Description, "current information", "web_search description should mention current information")

	require.NotNil(t, adoptionAdvisorSkill, "Should have adoption_advisor skill")
	require.Equal(t, "llm", adoptionAdvisorSkill.Type, "adoption_advisor should be llm type")
	require.Contains(t, adoptionAdvisorSkill.Description, "adoption and care", "adoption_advisor description should mention adoption and care")

	require.NotNil(t, knowledgeSearchSkill, "Should have knowledge_search skill")
	require.Equal(t, "nativeTool", knowledgeSearchSkill.Type, "knowledge_search should be nativeTool type")

	// Test metadata
	require.NotNil(t, agent.Metadata, "Metadata should not be nil")
	require.Equal(t, "1.0", agent.Metadata["version"], "Version should be 1.0")
	require.Equal(t, "Rescue dogs and animal welfare", agent.Metadata["specialization"], "Specialization should match")

	knowledgeService, err := knowledge.NewService(context.TODO(), &config.ModelConfig{
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}, config.NewKnowledgeConfig(), slog.Default())
	require.NoError(t, err)

	// Test runtime creation and execution with knowledge query
	runtime, err := agentruntime.NewAgentRuntime(
		context.TODO(),
		agentruntime.WithAgent(agent),
		agentruntime.WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
		agentruntime.WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
		agentruntime.WithKnowledgeService(knowledgeService),
		agentruntime.WithLogger(slog.Default()),
	)
	require.NoError(t, err)
	defer runtime.Close()

	var out string
	resp, err := runtime.Run(context.TODO(), engine.RunRequest{
		ThreadInstruction: "User asks about a rescue dog.",
		History: []engine.Conversation{
			{
				User: "USER",
				Text: "Can you tell me about Hosu? I heard he's also called Mandu sometimes.",
			},
		},
	}, &out)
	require.NoError(t, err)
	require.NotNil(t, resp)

	t.Logf("Response: %+v", resp)
	t.Logf("Output: %s", out)

	// Verify the output mentions Hosu
	require.NotEmpty(t, out, "Output should not be empty")
	// The agent should be able to reference the knowledge about Hosu
	// Check for any mention of the dog or related terms
	outputLower := strings.ToLower(out)
	hasRelevantContent := strings.Contains(outputLower, "hosu") ||
		strings.Contains(outputLower, "mandu") ||
		strings.Contains(outputLower, "rescue") ||
		strings.Contains(outputLower, "nuri")
	require.True(t, hasRelevantContent,
		"Output should mention Hosu, Mandu, rescue, or Nuri context")

	// Check if knowledge_search tool was called
	knowledgeSearchCalled := false
	for _, toolCall := range resp.ToolCalls {
		if toolCall.Name == "knowledge_search" {
			knowledgeSearchCalled = true
			t.Logf("Knowledge search tool was called with arguments: %s", string(toolCall.Arguments))
			t.Logf("Knowledge search results: %s", string(toolCall.Result))

			// Verify the tool call arguments contain relevant query
			require.Contains(t, strings.ToLower(string(toolCall.Arguments)), "hosu",
				"Knowledge search query should contain 'hosu'")
		}
	}

	// With the new tool-based approach, knowledge_search should be called
	require.True(t, knowledgeSearchCalled, "knowledge_search tool should have been called")

	// Additional verification for tool-based knowledge retrieval
	// Check if the response contains specific details from the knowledge base
	detailsFound := strings.Contains(outputLower, "3 years") ||
		strings.Contains(outputLower, "mixed breed") ||
		strings.Contains(outputLower, "gentle") ||
		strings.Contains(outputLower, "playful") ||
		strings.Contains(outputLower, "belly rubs") ||
		strings.Contains(outputLower, "seoul")

	t.Logf("Knowledge details found: %v", detailsFound)

	// Log a message about the new tool-based approach
	if knowledgeSearchCalled && detailsFound {
		t.Log("Tool-based knowledge retrieval is working correctly - agent called knowledge_search and used the results")
	} else if knowledgeSearchCalled && !detailsFound {
		t.Log("Warning: Agent called knowledge_search but may not have used the results effectively")
	} else {
		t.Log("Warning: Agent did not call knowledge_search tool - may need to update agent configuration")
	}
}

// TestAgentWithRAGAndCustomKnowledge tests RAG with more complex queries
func TestAgentWithRAGAndCustomKnowledge(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test because OPENAI_API_KEY is not set")
	}
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping test because ANTHROPIC_API_KEY is not set")
	}

	// Create a simple agent with knowledge
	agent := entity.Agent{
		AgentCard: entity.AgentCard{
			Name:        "TestRAGAgent",
			Description: "A test agent for RAG functionality",
		},
		ModelName: "anthropic/claude-4-sonnet",
		System:    "You are a helpful assistant with access to a knowledge base.",
		Role:      "Knowledge Assistant",
		Knowledge: []map[string]any{
			{
				"topic":   "Company Policy",
				"content": "Our company vacation policy allows 15 days of paid time off per year.",
				"details": "Vacation days must be approved by your manager at least 2 weeks in advance.",
			},
			{
				"topic":   "Office Hours",
				"content": "Office hours are Monday to Friday, 9 AM to 6 PM.",
				"details": "Remote work is allowed on Wednesdays and Fridays.",
			},
			{
				"topic":   "Health Benefits",
				"content": "Full health insurance coverage including dental and vision.",
				"details": "Coverage begins on the first day of employment. Family members can be added.",
			},
		},
		// Add knowledge_search skill for tool-based knowledge retrieval
		Skills: []entity.AgentSkill{
			{
				Name:        "knowledge_search",
				Type:        "nativeTool",
				Description: "Search through the knowledge base for relevant information",
			},
		},
	}

	// Create knowledge config with specific settings
	knowledgeConfig := config.NewKnowledgeConfig()
	knowledgeConfig.RerankEnabled = true
	knowledgeConfig.RerankTopK = 2
	knowledgeConfig.VectorEnabled = true

	knowledgeService, err := knowledge.NewService(context.TODO(), &config.ModelConfig{
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}, knowledgeConfig, slog.Default())
	require.NoError(t, err)

	runtime, err := agentruntime.NewAgentRuntime(
		context.TODO(),
		agentruntime.WithAgent(agent),
		agentruntime.WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
		agentruntime.WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
		agentruntime.WithKnowledgeService(knowledgeService),
		agentruntime.WithLogger(slog.Default()),
	)
	require.NoError(t, err)
	defer runtime.Close()

	// Test 1: Query about vacation policy
	var out1 string
	resp1, err := runtime.Run(context.TODO(), engine.RunRequest{
		History: []engine.Conversation{
			{
				User: "USER",
				Text: "How many vacation days do I get?",
			},
		},
	}, &out1)
	require.NoError(t, err)
	require.NotNil(t, resp1)

	t.Logf("Test 1 - Vacation Query Response: %s", out1)
	require.Contains(t, strings.ToLower(out1), "15 days", "Should mention 15 days of vacation")

	// Verify knowledge_search tool was called for vacation query
	vacationSearchCalled := false
	for _, toolCall := range resp1.ToolCalls {
		if toolCall.Name == "knowledge_search" {
			vacationSearchCalled = true
			t.Logf("Test 1 - Knowledge search called with: %s", string(toolCall.Arguments))
		}
	}
	require.True(t, vacationSearchCalled, "knowledge_search should be called for vacation query")

	// Test 2: Query about remote work
	var out2 string
	resp2, err := runtime.Run(context.TODO(), engine.RunRequest{
		History: []engine.Conversation{
			{
				User: "USER",
				Text: "Can I work from home?",
			},
		},
	}, &out2)
	require.NoError(t, err)
	require.NotNil(t, resp2)

	t.Logf("Test 2 - Remote Work Query Response: %s", out2)
	outputLower := strings.ToLower(out2)
	hasRemoteInfo := strings.Contains(outputLower, "wednesday") ||
		strings.Contains(outputLower, "friday") ||
		strings.Contains(outputLower, "remote")
	require.True(t, hasRemoteInfo, "Should mention remote work days")

	// Verify knowledge_search tool was called for remote work query
	remoteSearchCalled := false
	for _, toolCall := range resp2.ToolCalls {
		if toolCall.Name == "knowledge_search" {
			remoteSearchCalled = true
			t.Logf("Test 2 - Knowledge search called with: %s", string(toolCall.Arguments))
		}
	}
	require.True(t, remoteSearchCalled, "knowledge_search should be called for remote work query")

	// Test 3: Query about health benefits
	var out3 string
	resp3, err := runtime.Run(context.TODO(), engine.RunRequest{
		History: []engine.Conversation{
			{
				User: "USER",
				Text: "What health benefits are included?",
			},
		},
	}, &out3)
	require.NoError(t, err)
	require.NotNil(t, resp3)

	t.Logf("Test 3 - Health Benefits Query Response: %s", out3)
	outputLower3 := strings.ToLower(out3)
	hasHealthInfo := strings.Contains(outputLower3, "health insurance") ||
		strings.Contains(outputLower3, "dental") ||
		strings.Contains(outputLower3, "vision")
	require.True(t, hasHealthInfo, "Should mention health insurance details")

	// Verify knowledge_search tool was called for health benefits query
	healthSearchCalled := false
	for _, toolCall := range resp3.ToolCalls {
		if toolCall.Name == "knowledge_search" {
			healthSearchCalled = true
			t.Logf("Test 3 - Knowledge search called with: %s", string(toolCall.Arguments))
		}
	}
	require.True(t, healthSearchCalled, "knowledge_search should be called for health benefits query")

	t.Log("All tests passed - Tool-based knowledge retrieval is working correctly")
}

func TestAgentWithRAGAndPDFKnowledge(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test because OPENAI_API_KEY is not set")
	}
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping test because ANTHROPIC_API_KEY is not set")
	}

	ctx := context.TODO()
	pdfFile, err := os.ReadFile("./knowledge/testdata/solana-whitepaper-en.pdf")
	require.NoError(t, err)

	agentFile, err := os.ReadFile("./examples/solana_expert.agent.yaml")
	require.NoError(t, err)

	var agent entity.Agent
	err = yaml.Unmarshal(agentFile, &agent)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	knowledgeConfig := config.NewKnowledgeConfig()
	knowledgeConfig.SqlitePath = ":memory:"

	knowledgeService, err := knowledge.NewService(ctx, &config.ModelConfig{
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}, knowledgeConfig, logger)
	require.NoError(t, err)

	if _, err := knowledgeService.IndexKnowledgeFromPDF(ctx, "solana-whitepaper", bytes.NewReader(pdfFile)); err != nil {
		t.Fatalf("Failed to index knowledge from PDF: %v", err)
	}

	runtime, err := agentruntime.NewAgentRuntime(
		ctx,
		agentruntime.WithAgent(agent),
		agentruntime.WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
		agentruntime.WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
		agentruntime.WithKnowledgeService(knowledgeService),
		agentruntime.WithLogger(logger),
	)

	require.NoError(t, err)
	defer runtime.Close()

	var out string
	resp, err := runtime.Run(ctx, engine.RunRequest{
		History: []engine.Conversation{
			{
				User: "USER",
				Text: "What is Solana? Can you explain the details to me?",
			},
		},
	}, &out)
	require.NoError(t, err)
	require.NotNil(t, resp)

	t.Logf("Response: %+v", resp)
	t.Logf("Output: %s", out)

	require.True(t, strings.Contains(out, "Solana"), "Output should contain `Solana`")
}
