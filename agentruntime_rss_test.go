package agentruntime_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/habiliai/agentruntime"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRSSAgent(t *testing.T) {
	// Skip this test if ANTHROPIC_API_KEY is not set
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY is not set")
	}

	ctx := context.Background()

	// Load the startup news agent configuration
	configPath := "./examples/startup_news_agent.yaml"
	bytes, err := os.ReadFile(configPath)
	require.NoError(t, err, "Failed to read agent config file")

	var agent entity.Agent
	err = yaml.Unmarshal(bytes, &agent)
	require.NoError(t, err, "Failed to unmarshal agent config")

	// Verify agent configuration
	require.Equal(t, "TechInsight", agent.Name)
	require.Contains(t, agent.Description, "startup news specialist")
	require.Len(t, agent.Skills, 1, "Should have 1 skill")
	require.Equal(t, "nativeTool", agent.Skills[0].Type)
	require.NotNil(t, agent.Skills[0].OfNative)
	require.Equal(t, "rss", agent.Skills[0].OfNative.Name)

	// Create agent runtime
	runtime, err := agentruntime.NewAgentRuntime(
		ctx,
		agentruntime.WithAgent(agent),
		agentruntime.WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
	)
	require.NoError(t, err, "Failed to create agent runtime")
	defer runtime.Close()

	// Test cases
	testCases := []struct {
		name     string
		message  string
		validate func(t *testing.T, response *engine.RunResponse, output string)
	}{
		{
			name:    "Search for AI startup news",
			message: "What are the latest news about AI startups?",
			validate: func(t *testing.T, response *engine.RunResponse, output string) {
				assert.NotEmpty(t, output)
				// The agent should use the search_rss tool
				assert.NotEmpty(t, response.ToolCalls)
				fmt.Printf("AI Startup News Response:\n%s\n", output)
			},
		},
		{
			name:    "Read TechCrunch feed",
			message: "Show me the latest 5 articles from TechCrunch startups feed",
			validate: func(t *testing.T, response *engine.RunResponse, output string) {
				assert.NotEmpty(t, output)
				// The agent should use the read_rss tool
				assert.NotEmpty(t, response.ToolCalls)
				fmt.Printf("TechCrunch Feed Response:\n%s\n", output)
			},
		},
		{
			name:    "Search for funding news",
			message: "Find recent startup funding announcements",
			validate: func(t *testing.T, response *engine.RunResponse, output string) {
				assert.NotEmpty(t, output)
				assert.NotEmpty(t, response.ToolCalls)
				fmt.Printf("Funding News Response:\n%s\n", output)
			},
		},
		{
			name:    "Search across multiple feeds",
			message: "Search for 'unicorn' startups across all available news feeds",
			validate: func(t *testing.T, response *engine.RunResponse, output string) {
				assert.NotEmpty(t, output)
				assert.NotEmpty(t, response.ToolCalls)
				fmt.Printf("Unicorn Search Response:\n%s\n", output)
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var output string
			response, err := runtime.Run(ctx, engine.RunRequest{
				ThreadInstruction: "User is asking about startup news and trends.",
				History: []engine.Conversation{
					{
						User: "USER",
						Text: tc.message,
					},
				},
			}, &output)

			require.NoError(t, err, "Failed to run agent")
			require.NotNil(t, response)

			// Validate response
			tc.validate(t, response, output)

			// Print separator for clarity
			fmt.Println("\n" + strings.Repeat("-", 80) + "\n")
		})
	}
}

// TestRSSAgentToolCalls specifically tests the RSS tool calls
func TestRSSAgentToolCalls(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY is not set")
	}

	ctx := context.Background()

	// Load the agent configuration
	bytes, err := os.ReadFile("./examples/startup_news_agent.yaml")
	require.NoError(t, err)

	var agent entity.Agent
	err = yaml.Unmarshal(bytes, &agent)
	require.NoError(t, err)

	// Create agent runtime
	runtime, err := agentruntime.NewAgentRuntime(
		ctx,
		agentruntime.WithAgent(agent),
		agentruntime.WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
	)
	require.NoError(t, err)
	defer runtime.Close()

	// Test search_rss tool
	t.Run("search_rss tool call", func(t *testing.T) {
		var output string
		response, err := runtime.Run(ctx, engine.RunRequest{
			ThreadInstruction: "User is looking for specific startup news.",
			History: []engine.Conversation{
				{
					User: "USER",
					Text: "Search for 'Series A' funding news from Crunchbase and TechCrunch feeds",
				},
			},
		}, &output)

		require.NoError(t, err)
		require.NotNil(t, response)
		require.NotEmpty(t, response.ToolCalls)

		// Check if search_rss was called
		searchRSSCalled := false
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "search_rss" {
				searchRSSCalled = true
				t.Logf("search_rss called with args: %s", string(toolCall.Arguments))

				// Verify the arguments contain the query
				arguments := strings.ToLower(string(toolCall.Arguments))
				if !strings.Contains(arguments, "series a") &&
					!strings.Contains(arguments, "startup") &&
					!strings.Contains(arguments, "funding") &&
					!strings.Contains(arguments, "crunchbase") &&
					!strings.Contains(arguments, "techcrunch") {
					t.Fatalf("search_rss called with args: %s", arguments)
				}
			}
		}

		assert.True(t, searchRSSCalled, "search_rss tool should have been called")
		t.Logf("Series A Funding Search:\n%s\n", output)
	})
}

// TestRSSAgentErrorHandling tests error handling scenarios
func TestRSSAgentErrorHandling(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY is not set")
	}

	ctx := context.Background()

	// Load the agent configuration
	bytes, err := os.ReadFile("./examples/startup_news_agent.yaml")
	require.NoError(t, err)

	var agent entity.Agent
	err = yaml.Unmarshal(bytes, &agent)
	require.NoError(t, err)

	// Create agent runtime
	runtime, err := agentruntime.NewAgentRuntime(
		ctx,
		agentruntime.WithAgent(agent),
		agentruntime.WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
	)
	require.NoError(t, err)
	defer runtime.Close()

	// Test with a query that might not return results
	var output string
	response, err := runtime.Run(ctx, engine.RunRequest{
		ThreadInstruction: "User is searching for very specific news.",
		History: []engine.Conversation{
			{
				User: "USER",
				Text: "Find news about 'zyxwvutsrqponmlkjihgfedcba' startups",
			},
		},
	}, &output)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Even with no results, the agent should respond appropriately
	assert.NotEmpty(t, output)
}
