package agentruntime_test

import (
	"context"
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
