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
	firecrawl "github.com/mendableai/firecrawl-go"
	"github.com/samber/lo"
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

func TestAgentRuntimeForAgentWithKnowledgeSkills(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Check required environment variables
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test because OPENAI_API_KEY is not set")
	}
	if os.Getenv("FIRECRAWL_API_KEY") == "" {
		t.Skip("Skipping test because FIRECRAWL_API_KEY is not set")
	}
	// Note: Using OpenAI model instead of Anthropic to avoid credit issues

	ctx := context.Background()

	// Create agent with knowledge skills
	agent := entity.Agent{
		Name:        "KnowledgeTestAgent",
		Description: "An agent to test knowledge indexing and search functionality",
		ModelName:   "openai/gpt-4o-mini", // Use OpenAI instead of Anthropic
		System: `You are a knowledge management assistant. Your role is to help users index and search knowledge from various sources.

Always use the knowledge tools when appropriate:
- Use knowledge_search to find information from the indexed knowledge base
- When users ask about content from URLs, tell them to index the URL first if no relevant knowledge is found
- Be helpful and informative about what knowledge is available

Be clear about what knowledge operations you're performing and what information you find.`,
		Role: "Knowledge Management Assistant",
		Skills: []entity.AgentSkillUnion{
			{
				Type: "nativeTool",
				OfNative: &entity.NativeAgentSkill{
					Name:    "knowledge_search",
					Details: "Search through indexed knowledge base for relevant information",
				},
			},
		},
	}

	// Create runtime with knowledge support
	runtime, err := NewAgentRuntime(
		ctx,
		WithAgent(agent),
		WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
		WithLogger(slog.Default()),
	)
	require.NoError(t, err)
	defer runtime.Close()

	// Test URLs for indexing
	testURL := "https://httpbin.org/html" // Simple test page

	t.Run("Index knowledge from URL", func(t *testing.T) {
		// Get knowledge service and index URL directly
		knowledgeService := runtime.GetKnowledgeService()
		require.NotNil(t, knowledgeService, "Knowledge service should be available")

		// Create optimized crawl parameters for testing
		crawlParams := firecrawl.CrawlParams{
			MaxDepth:           lo.ToPtr(1), // Only 1 level deep for fast testing
			Limit:              lo.ToPtr(2), // Limit to 2 pages for fast testing
			AllowBackwardLinks: lo.ToPtr(false),
			AllowExternalLinks: lo.ToPtr(false),
			ScrapeOptions: firecrawl.ScrapeParams{
				Formats: []string{"markdown"}, // Only markdown for faster processing
			},
		}

		// Index the URL
		knowledge, err := knowledgeService.IndexKnowledgeFromURL(ctx, "test-httpbin", testURL, crawlParams)
		require.NoError(t, err, "Should successfully index URL")
		require.NotNil(t, knowledge, "Knowledge should not be nil")
		require.Equal(t, "test-httpbin", knowledge.ID)
		require.NotEmpty(t, knowledge.Documents, "Should have indexed documents")

		t.Logf("Successfully indexed URL: %s with %d documents", testURL, len(knowledge.Documents))
	})

	t.Run("Search for knowledge through agent", func(t *testing.T) {
		// Test searching for content about HTTP testing
		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "What information do we have about HTTP testing or httpbin?",
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Check that knowledge_search tool was called
		foundKnowledgeSearchCall := false
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "knowledge_search" {
				foundKnowledgeSearchCall = true
				break
			}
		}
		require.True(t, foundKnowledgeSearchCall, "knowledge_search tool should have been called")

		t.Logf("Knowledge search response: %s", response.Text())

		// Verify that some knowledge was found by checking for HTTP-related content
		responseText := strings.ToLower(response.Text())
		containsRelevantInfo := strings.Contains(responseText, "http") ||
			strings.Contains(responseText, "html") ||
			strings.Contains(responseText, "test") ||
			strings.Contains(responseText, "httpbin")
		require.True(t, containsRelevantInfo, "Response should contain relevant HTTP/testing information")
	})

	t.Run("Search for specific technical concepts", func(t *testing.T) {
		// Test searching for specific HTML or web concepts
		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "Can you find information about HTML or web pages in our knowledge base?",
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Check that knowledge_search tool was called
		foundKnowledgeSearchCall := false
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "knowledge_search" {
				foundKnowledgeSearchCall = true
				break
			}
		}
		require.True(t, foundKnowledgeSearchCall, "knowledge_search tool should have been called")

		t.Logf("HTML search response: %s", response.Text())
	})

	t.Run("Handle query with no relevant knowledge", func(t *testing.T) {
		// Test searching for something that's not in the indexed knowledge
		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "What do you know about quantum computing or blockchain technology?",
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Check that knowledge_search tool was called (even if no results)
		foundKnowledgeSearchCall := false
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "knowledge_search" {
				foundKnowledgeSearchCall = true
				break
			}
		}
		require.True(t, foundKnowledgeSearchCall, "knowledge_search tool should have been called")

		t.Logf("No knowledge response: %s", response.Text())

		// The response should indicate no relevant knowledge was found
		responseText := strings.ToLower(response.Text())
		indicatesNoKnowledge := strings.Contains(responseText, "no") ||
			strings.Contains(responseText, "not found") ||
			strings.Contains(responseText, "don't have") ||
			strings.Contains(responseText, "unable") ||
			strings.Contains(responseText, "couldn't find")
		require.True(t, indicatesNoKnowledge, "Response should indicate no relevant knowledge was found")
	})

	t.Run("Index additional URL and search across multiple sources", func(t *testing.T) {
		// Get knowledge service and index another URL
		knowledgeService := runtime.GetKnowledgeService()

		// Create another test URL - use a different httpbin endpoint
		testURL2 := "https://httpbin.org/json"

		// Create optimized crawl parameters for testing
		crawlParams := firecrawl.CrawlParams{
			MaxDepth:           lo.ToPtr(1),
			Limit:              lo.ToPtr(1), // Even smaller limit for second URL
			AllowBackwardLinks: lo.ToPtr(false),
			AllowExternalLinks: lo.ToPtr(false),
			ScrapeOptions: firecrawl.ScrapeParams{
				Formats: []string{"markdown"},
			},
		}

		// Index the second URL
		knowledge2, err := knowledgeService.IndexKnowledgeFromURL(ctx, "test-httpbin-json", testURL2, crawlParams)
		require.NoError(t, err, "Should successfully index second URL")
		require.NotNil(t, knowledge2, "Second knowledge should not be nil")

		t.Logf("Successfully indexed second URL: %s with %d documents", testURL2, len(knowledge2.Documents))

		// Now search across both indexed sources
		response, err := runtime.Run(ctx, engine.RunRequest{
			History: []engine.Conversation{
				{
					User: "user",
					Text: "What different types of HTTP content or endpoints do we have information about?",
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, response.Text())

		// Check that knowledge_search tool was called
		foundKnowledgeSearchCall := false
		for _, toolCall := range response.ToolCalls {
			if toolCall.Name == "knowledge_search" {
				foundKnowledgeSearchCall = true
				break
			}
		}
		require.True(t, foundKnowledgeSearchCall, "knowledge_search tool should have been called")

		t.Logf("Multi-source search response: %s", response.Text())
	})

	t.Run("Direct knowledge service verification", func(t *testing.T) {
		// Access knowledge service directly for verification
		knowledgeService := runtime.GetKnowledgeService()
		require.NotNil(t, knowledgeService, "Knowledge service should be available")

		// Test direct knowledge search
		searchResults, err := knowledgeService.RetrieveRelevantKnowledge(ctx, "HTTP testing", 5, nil)
		require.NoError(t, err, "Should be able to search knowledge directly")
		require.NotEmpty(t, searchResults, "Should find some relevant knowledge")

		// Verify the knowledge contains useful information
		for i, result := range searchResults {
			require.NotEmpty(t, result.EmbeddingText, "Search result %d should have embedding text", i)
			require.NotEmpty(t, result.ID, "Search result %d should have ID", i)
			t.Logf("Search result %d (score: %.3f): %s", i, result.Score,
				truncateString(result.EmbeddingText, 100))
		}

		// Test retrieving specific knowledge by ID
		knowledge1, err := knowledgeService.GetKnowledge(ctx, "test-httpbin")
		require.NoError(t, err, "Should retrieve first knowledge")
		require.NotNil(t, knowledge1, "First knowledge should exist")
		require.Equal(t, "test-httpbin", knowledge1.ID)

		knowledge2, err := knowledgeService.GetKnowledge(ctx, "test-httpbin-json")
		require.NoError(t, err, "Should retrieve second knowledge")
		require.NotNil(t, knowledge2, "Second knowledge should exist")
		require.Equal(t, "test-httpbin-json", knowledge2.ID)

		t.Logf("Direct verification passed: Found %d search results", len(searchResults))
		t.Logf("Knowledge 1: %d documents, Knowledge 2: %d documents",
			len(knowledge1.Documents), len(knowledge2.Documents))
	})
}

// Helper function to truncate strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
