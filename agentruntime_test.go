package agentruntime_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/habiliai/agentruntime"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/require"
)

func TestAgentRuntime(t *testing.T) {
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
