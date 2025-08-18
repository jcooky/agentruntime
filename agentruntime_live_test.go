package agentruntime

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/require"
)

func TestAgentRuntimeForAgentWithMemorySkills(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Check required environment variables [[memory:3743077]]
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test because OPENAI_API_KEY is not set")
	}
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping test because ANTHROPIC_API_KEY is not set")
	}

	ctx := context.Background()

	// Create agent with memory skills
	agent := entity.Agent{
		Name:        "MemoryTestAgent",
		Description: "An agent to test memory skills functionality",
		ModelName:   "anthropic/claude-3.5-haiku",
		System: `You are a memory test assistant. Your role is to demonstrate memory functionality.
Always use the memory tools when appropriate:
- Store user information immediately when shared
- Search for relevant memories when needed
- Recall specific memories by key when possible
- List all memories when requested
- Delete memories when asked

Be clear about what memory operations you're performing.`,
		Role: "Memory Test Assistant",
		Skills: []entity.AgentSkillUnion{
			{
				Type: "nativeTool",
				OfNative: &entity.NativeAgentSkill{
					Name:    "memory",
					Details: "Memory management tools for storing and retrieving information",
				},
			},
		},
	}

	// Create runtime with memory support
	runtime, err := NewAgentRuntime(
		ctx,
		WithAgent(agent),
		WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
		WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
		WithLogger(slog.Default()),
	)
	require.NoError(t, err)
	defer runtime.Close()

	t.Run("Store and retrieve user information", func(t *testing.T) {
		// Test storing personal information
		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "Hi! My name is Dennis and I'm a software engineer at HabiliAI. I love drinking coffee, especially dark roast.",
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Check that remember_memory tool was called (more reliable than text checking)
		foundRememberMemoryCall := false
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "remember_memory" {
				foundRememberMemoryCall = true
				break
			}
		}
		require.True(t, foundRememberMemoryCall, "remember_memory tool should have been called to store user information")

		// Additionally, verify that multiple memory items were stored
		rememberCallCount := 0
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "remember_memory" {
				rememberCallCount++
			}
		}
		require.GreaterOrEqual(t, rememberCallCount, 2, "Should have stored multiple pieces of information (name, job, preferences, etc.)")

		t.Logf("Store response: %s", response.Text())
		t.Logf("Tool calls made: %d remember_memory calls", rememberCallCount)
	})

	t.Run("Search for stored memories", func(t *testing.T) {
		// Test searching for coffee preferences
		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "What do you know about my coffee preferences?",
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Verify search_memory tool was called (more reliable than text checking)
		foundSearchMemoryCall := false
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "search_memory" {
				foundSearchMemoryCall = true
				break
			}
		}
		require.True(t, foundSearchMemoryCall, "search_memory tool should have been called")

		// Verify that coffee preference exists in memory via direct service check
		memoryService := runtime.GetMemoryService()
		memories, err := memoryService.ListMemories(ctx)
		require.NoError(t, err)

		coffeeMemoryExists := false
		for _, memory := range memories {
			if strings.Contains(strings.ToLower(memory.Value), "coffee") {
				coffeeMemoryExists = true
				break
			}
		}
		require.True(t, coffeeMemoryExists, "Coffee preference should exist in memory")

		t.Logf("Search response: %s", response.Text())
	})

	t.Run("List all stored memories", func(t *testing.T) {
		// Test listing all memories
		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "Can you show me everything you remember about me?",
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Verify list_memories tool was called
		foundListMemoriesCall := false
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "list_memories" {
				foundListMemoriesCall = true
				break
			}
		}
		require.True(t, foundListMemoriesCall, "list_memories tool should have been called")

		// Verify essential information exists via direct service check
		memoryService := runtime.GetMemoryService()
		memories, err := memoryService.ListMemories(ctx)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(memories), 3, "Should have at least 3 memories stored")

		t.Logf("List memories response: %s", response.Text())
	})

	t.Run("Store additional information and search", func(t *testing.T) {
		// Store additional information
		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "I also want to mention that I'm working on an AI agent runtime project and my favorite programming language is Go.",
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Verify remember_memory tool was called for new info
		rememberCallCount := 0
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "remember_memory" {
				rememberCallCount++
			}
		}
		require.GreaterOrEqual(t, rememberCallCount, 1, "Should have stored new information")

		t.Logf("Additional info response: %s", response.Text())

		// Search for project information
		response, err = runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "What project am I working on?",
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Verify search tools were called
		searchToolCalled := false
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "search_memory" || toolCall.Name == "list_memories" {
				searchToolCalled = true
				break
			}
		}
		require.True(t, searchToolCalled, "Should have used memory tools to search for project info")

		t.Logf("Project search response: %s", response.Text())
	})

	t.Run("Test memory deletion", func(t *testing.T) {
		// Ask to delete specific memory
		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "Please delete any information about my coffee preferences. I want to forget that.",
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Verify delete_memory tool was called
		foundDeleteMemoryCall := false
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "delete_memory" {
				foundDeleteMemoryCall = true
				break
			}
		}
		require.True(t, foundDeleteMemoryCall, "delete_memory tool should have been called")

		t.Logf("Delete memory response: %s", response.Text())

		// Verify the coffee preference is deleted by checking memory service directly
		response, err = runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "What do you know about my coffee preferences now?",
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Check memory service directly - coffee preference should be gone
		memoryService := runtime.GetMemoryService()
		memories, err := memoryService.ListMemories(ctx)
		require.NoError(t, err)

		coffeeMemoryExists := false
		for _, memory := range memories {
			if strings.Contains(strings.ToLower(memory.Value), "coffee") && strings.Contains(strings.ToLower(memory.Value), "dark roast") {
				coffeeMemoryExists = true
				break
			}
		}
		require.False(t, coffeeMemoryExists, "Coffee preference should be deleted from memory")

		t.Logf("Verify deletion response: %s", response.Text())
	})

	t.Run("Final memory state verification", func(t *testing.T) {
		// Check final state of memories
		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "What do you still remember about me after the deletion?",
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Verify memory tools were used
		memoryToolCalled := false
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "list_memories" || toolCall.Name == "search_memory" {
				memoryToolCalled = true
				break
			}
		}
		require.True(t, memoryToolCalled, "Should have used memory tools to check final state")

		t.Logf("Final state response: %s", response.Text())
	})

	// The most reliable test: Direct memory service verification
	t.Run("Direct memory service verification", func(t *testing.T) {
		// Access memory service directly for most reliable verification
		memoryService := runtime.GetMemoryService()
		require.NotNil(t, memoryService, "Memory service should be available")

		// List all memories directly from the service
		memories, err := memoryService.ListMemories(ctx)
		require.NoError(t, err)

		// Create a map of actual memory keys for verification
		memoryKeys := make(map[string]bool)
		memoryValues := make(map[string]string)
		for _, memory := range memories {
			memoryKeys[memory.Key] = true
			memoryValues[memory.Key] = memory.Value
			t.Logf("Found memory: %s = %s (tags: %v)", memory.Key, memory.Value, memory.Tags)
		}

		// Verify essential memories exist (using actual keys from the logs)
		require.True(t, memoryKeys["user_name_full"], "Should have stored user name")

		// Check for job info (could be stored as separate title/company or combined)
		hasJobInfo := memoryKeys["user_job_title"] || memoryKeys["user_job_company"] ||
			(len(memoryValues) > 0 && (strings.Contains(strings.ToLower(fmt.Sprintf("%v", memoryValues)), "engineer") ||
				strings.Contains(strings.ToLower(fmt.Sprintf("%v", memoryValues)), "habili")))
		require.True(t, hasJobInfo, "Should have stored job information")

		// Check for project info (could have different key name)
		hasProjectInfo := false
		for key, value := range memoryValues {
			if strings.Contains(strings.ToLower(key), "project") || strings.Contains(strings.ToLower(value), "ai agent") || strings.Contains(strings.ToLower(value), "runtime") {
				hasProjectInfo = true
				break
			}
		}
		require.True(t, hasProjectInfo, "Should have stored project info")

		// Check for programming language preference
		hasProgrammingLanguage := false
		for key, value := range memoryValues {
			if strings.Contains(strings.ToLower(key), "programming") || strings.Contains(strings.ToLower(value), "go") {
				hasProgrammingLanguage = true
				break
			}
		}
		require.True(t, hasProgrammingLanguage, "Should have stored programming language preference")

		// Verify coffee preferences are deleted
		coffeeMemoryExists := false
		for _, value := range memoryValues {
			if strings.Contains(strings.ToLower(value), "coffee") && strings.Contains(strings.ToLower(value), "dark roast") {
				coffeeMemoryExists = true
				break
			}
		}
		require.False(t, coffeeMemoryExists, "Should NOT have coffee preferences (deleted)")

		// Verify specific memory content for name
		if memoryKeys["user_name_full"] {
			nameMemory, err := memoryService.GetMemory(ctx, "user_name_full")
			require.NoError(t, err)
			require.Equal(t, "Dennis", nameMemory.Value)
			require.Contains(t, nameMemory.Tags, "personal")
		}

		t.Logf("Direct verification passed: Found %d memories", len(memories))
	})
}

func TestAgentRuntimeWithUserInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Check required environment variables
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test because OPENAI_API_KEY is not set")
	}
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping test because ANTHROPIC_API_KEY is not set")
	}

	ctx := context.Background()

	// Create agent that can use user info
	agent := entity.Agent{
		Name:        "PersonalizedAgent",
		Description: "An agent that provides personalized responses based on user information",
		ModelName:   "openai/gpt-4o-mini",
		System: `You are a personalized assistant. Use the provided user information to tailor your responses.
When user information is available:
- Address the user by their name when appropriate
- Consider their location, company, and interests in your responses
- Be more helpful by understanding their context
- Reference their background when relevant

Always be professional but friendly, and make good use of the user context provided.`,
		Role: "Personalized Assistant",
		Skills: []entity.AgentSkillUnion{
			{
				Type: "nativeTool",
				OfNative: &entity.NativeAgentSkill{
					Name:    "memory",
					Details: "Memory management tools for enhanced personalization",
				},
			},
		},
	}

	// Create runtime
	runtime, err := NewAgentRuntime(
		ctx,
		WithAgent(agent),
		WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
		WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
		WithLogger(slog.Default()),
	)
	require.NoError(t, err)
	defer runtime.Close()

	t.Run("Test with complete user info", func(t *testing.T) {
		// Test with full user information
		userInfo := &engine.UserInfo{
			FullName: "Dennis Kim",
			Username: "dennis",
			Location: "Seoul, South Korea",
			Company:  "HabiliAI",
			Bio:      "Software engineer passionate about AI and automation",
		}

		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "Hi! Could you help me plan my day? I have some coding work to do.",
				},
			},
			UserInfo: userInfo,
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		responseText := strings.ToLower(response.Text())

		// Verify that the response uses user information contextually
		hasPersonalization := strings.Contains(responseText, "dennis") ||
			strings.Contains(responseText, "seoul") ||
			strings.Contains(responseText, "habili") ||
			strings.Contains(responseText, "korea")

		require.True(t, hasPersonalization, "Response should use user information for personalization")

		t.Logf("Complete user info response: %s", response.Text())
	})

	t.Run("Test with partial user info", func(t *testing.T) {
		// Test with minimal user information
		userInfo := &engine.UserInfo{
			FullName: "Dennis",
			Company:  "HabiliAI",
		}

		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "What's the best approach for implementing AI agents in a startup environment?",
				},
			},
			UserInfo: userInfo,
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		responseText := strings.ToLower(response.Text())

		// Should use the available user context
		hasPersonalization := strings.Contains(responseText, "dennis") ||
			strings.Contains(responseText, "habili") ||
			strings.Contains(responseText, "startup")

		require.True(t, hasPersonalization, "Response should use available user information")

		t.Logf("Partial user info response: %s", response.Text())
	})

	t.Run("Test without user info", func(t *testing.T) {
		// Test without user information (should still work)
		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "Tell me about Go programming best practices.",
				},
			},
			UserInfo: nil, // No user info provided
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Should still provide a helpful response even without user context
		require.Contains(t, strings.ToLower(response.Text()), "go", "Should provide Go programming advice")

		t.Logf("No user info response: %s", response.Text())
	})

	t.Run("Test user info with memory integration", func(t *testing.T) {
		// Test that user info and memory work well together
		userInfo := &engine.UserInfo{
			FullName: "Dennis Kim",
			Username: "dennis",
			Location: "Seoul, South Korea",
			Company:  "HabiliAI",
			Bio:      "AI engineer working on agent runtime systems",
		}

		// First interaction: introduce a preference
		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "I prefer TypeScript over JavaScript for frontend development. Please remember this.",
				},
			},
			UserInfo: userInfo,
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Check if memory tool was used
		foundMemoryCall := false
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "remember_memory" {
				foundMemoryCall = true
				break
			}
		}
		require.True(t, foundMemoryCall, "Should store user preference in memory")

		t.Logf("Preference storage response: %s", response.Text())

		// Second interaction: ask for personalized advice
		response, err = runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "I'm building a new web application. What technology stack would you recommend?",
				},
			},
			UserInfo: userInfo,
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		responseText := strings.ToLower(response.Text())

		// Should use both user info and memory for personalized recommendation
		hasPersonalization := strings.Contains(responseText, "typescript") ||
			strings.Contains(responseText, "dennis") ||
			(strings.Contains(responseText, "seoul") || strings.Contains(responseText, "korea"))

		require.True(t, hasPersonalization, "Should provide personalized recommendation using both user info and memory")

		t.Logf("Personalized recommendation response: %s", response.Text())
	})
}
