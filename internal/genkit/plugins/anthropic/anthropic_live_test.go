package anthropic_test

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/anthropic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/red_image.png
var redImageFile []byte

func TestLive_GenerateText(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	ctx := context.Background()
	g, err := genkit.Init(ctx, genkit.WithPlugins(&anthropic.Plugin{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}))
	require.NoError(t, err)

	tests := []struct {
		name      string
		modelName string
		prompt    string
		timeout   time.Duration
	}{
		{
			name:      "claude-3.5-haiku simple",
			modelName: "claude-3.5-haiku",
			prompt:    "What is 2+2? Answer with just the number.",
			timeout:   30 * time.Minute,
		},
		{
			name:      "claude-3.7-sonnet simple",
			modelName: "claude-3.7-sonnet",
			prompt:    "What is the capital of Japan? Answer with just the city name.",
			timeout:   30 * time.Minute,
		},
		{
			name:      "claude-4-sonnet simple",
			modelName: "claude-4-sonnet",
			prompt:    "What is the capital of France? Answer with just the city name.",
			timeout:   30 * time.Minute,
		},
		{
			name:      "claude-3.5-haiku with math",
			modelName: "claude-3.5-haiku",
			prompt:    "What is 10 divided by 2? Answer with just the number.",
			timeout:   30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := anthropic.Model(g, tt.modelName)
			if model == nil {
				t.Skipf("Model %s not available", tt.modelName)
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeoutCause(
				context.Background(),
				tt.timeout,
				fmt.Errorf("timeout %s", tt.modelName),
			)
			defer cancel()

			req := &ai.ModelRequest{
				Messages: []*ai.Message{
					{
						Role: ai.RoleUser,
						Content: []*ai.Part{
							ai.NewTextPart(tt.prompt),
						},
					},
				},
				Config: &ai.GenerationCommonConfig{
					MaxOutputTokens: 100,
					Temperature:     0.0,
				},
			}

			resp, err := model.Generate(ctx, req, nil)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Message)
			require.NotEmpty(t, resp.Message.Content)

			content := resp.Message.Content[0].Text
			assert.NotEmpty(t, content)
			fmt.Printf("%s response: %s\n", tt.modelName, content)
		})
	}
}

func TestLive_GenerateWithStreaming(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	ctx := context.TODO()
	g, err := genkit.Init(ctx, genkit.WithPlugins(&anthropic.Plugin{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}))
	require.NoError(t, err)

	model := anthropic.Model(g, "claude-3.5-haiku")
	require.NotNil(t, model)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	req := &ai.ModelRequest{
		Messages: []*ai.Message{
			{
				Role: ai.RoleUser,
				Content: []*ai.Part{
					ai.NewTextPart("Count from 1 to 5. Just the numbers, one per line."),
				},
			},
		},
		Config: &ai.GenerationCommonConfig{
			MaxOutputTokens: 50,
		},
	}

	var chunks []string
	resp, err := model.Generate(ctx, req, func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
		if len(chunk.Content) > 0 && chunk.Content[0].Text != "" {
			chunks = append(chunks, chunk.Content[0].Text)
			fmt.Print(chunk.Content[0].Text)
		}
		return nil
	})
	fmt.Println()

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, chunks)
	assert.NotEmpty(t, resp.Message.Content[0].Text)
}

func TestLive_GenerateWithImage(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	ctx := context.Background()
	g, err := genkit.Init(ctx, genkit.WithPlugins(&anthropic.Plugin{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}))
	require.NoError(t, err)

	model := anthropic.Model(g, "claude-3.5-haiku")
	require.NotNil(t, model)

	// Create a simple base64 encoded image (200x200 red pixel PNG)
	base64Image := base64.StdEncoding.EncodeToString(redImageFile)

	req := &ai.ModelRequest{
		Messages: []*ai.Message{
			{
				Role: ai.RoleUser,
				Content: []*ai.Part{
					ai.NewTextPart("What color is this image? Answer with just the color name."),
					ai.NewMediaPart("image/png", base64Image),
				},
			},
		},
		Config: &ai.GenerationCommonConfig{
			MaxOutputTokens: 50,
			Temperature:     0.0,
		},
	}

	resp, err := model.Generate(ctx, req, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Message)

	content := resp.Message.Content[0].Text
	assert.Contains(t, strings.ToLower(content), "red", "Expected the model to identify the red pixel")
	fmt.Printf("Image analysis response: %s\n", content)
}

func TestLive_GenerateWithSystemMessage(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	ctx := context.TODO()
	g, err := genkit.Init(ctx, genkit.WithPlugins(&anthropic.Plugin{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}))
	require.NoError(t, err)

	model := anthropic.Model(g, "claude-3.5-haiku")
	require.NotNil(t, model)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	req := &ai.ModelRequest{
		Messages: []*ai.Message{
			{
				Role: ai.RoleSystem,
				Content: []*ai.Part{
					ai.NewTextPart("You are a helpful assistant that always responds in haiku format."),
				},
			},
			{
				Role: ai.RoleUser,
				Content: []*ai.Part{
					ai.NewTextPart("Tell me about the ocean."),
				},
			},
		},
		Config: &ai.GenerationCommonConfig{
			MaxOutputTokens: 100,
		},
	}

	resp, err := model.Generate(ctx, req, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	content := resp.Message.Content[0].Text
	assert.NotEmpty(t, content)
	fmt.Printf("Haiku response: %s\n", content)
}

// TestLive_GenerateWithReasoning tests the extended thinking (reasoning) feature
// which allows models to show their step-by-step thought process.
// This feature is especially useful for complex problems that require logical reasoning.
func TestLive_GenerateWithReasoning(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	ctx := context.Background()
	g, err := genkit.Init(ctx, genkit.WithPlugins(&anthropic.Plugin{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}))
	require.NoError(t, err)

	tests := []struct {
		name        string
		modelName   string
		prompt      string
		enabled     bool
		budgetRatio float64
		timeout     time.Duration
	}{
		{
			name:        "claude-3.5-haiku without reasoning",
			modelName:   "claude-3.5-haiku",
			prompt:      "If a train travels 120 miles in 2 hours, what is its speed in mph? Think step by step.",
			enabled:     false,
			budgetRatio: 0,
			timeout:     30 * time.Minute,
		},
		{
			name:        "claude-3.7-sonnet with reasoning enabled",
			modelName:   "claude-3.7-sonnet",
			prompt:      "A farmer has 17 sheep. All but 9 die. How many are left? Think step by step.",
			enabled:     true,
			budgetRatio: 0.25, // 25% of maxOutputTokens
			timeout:     30 * time.Minute,
		},
		{
			name:        "claude-4-sonnet with reasoning enabled",
			modelName:   "claude-4-sonnet",
			prompt:      "A bat and a ball cost $1.10 in total. The bat costs $1.00 more than the ball. How much does the ball cost? Think carefully.",
			enabled:     true,
			budgetRatio: 0.25, // 25% of maxOutputTokens
			timeout:     30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := anthropic.Model(g, tt.modelName)
			if model == nil {
				t.Skipf("Model %s not available", tt.modelName)
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeoutCause(
				context.Background(),
				tt.timeout,
				fmt.Errorf("timeout %s", tt.modelName),
			)
			defer cancel()

			config := struct {
				ai.GenerationCommonConfig
				anthropic.ExtendedThinkingConfig
			}{
				GenerationCommonConfig: ai.GenerationCommonConfig{
					MaxOutputTokens: 5000, // Large enough to support reasoning budget (5000 * 0.25 = 1250 > 1024)
					Temperature:     0.0,
				},
				ExtendedThinkingConfig: anthropic.ExtendedThinkingConfig{
					ExtendedThinkingEnabled:     tt.enabled,
					ExtendedThinkingBudgetRatio: tt.budgetRatio,
				},
			}

			req := &ai.ModelRequest{
				Messages: []*ai.Message{
					{
						Role: ai.RoleUser,
						Content: []*ai.Part{
							ai.NewTextPart(tt.prompt),
						},
					},
				},
				Config: config,
			}

			resp, err := model.Generate(ctx, req, nil)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Message)
			require.NotEmpty(t, resp.Message.Content)

			// Check for reasoning part
			hasReasoning := false
			var reasoningText string
			var finalAnswer string

			for _, part := range resp.Message.Content {
				if part.IsReasoning() {
					hasReasoning = true
					reasoningText = part.Text
					t.Logf("Found reasoning part with %d characters", len(reasoningText))
				} else if part.IsText() {
					finalAnswer = part.Text
				}
			}

			// claude-3.5-haiku doesn't support reasoning, so skip assertions for it
			if tt.modelName != "claude-3.5-haiku" {
				if tt.enabled && tt.budgetRatio > 0 {
					assert.True(t, hasReasoning, "Expected reasoning part to be present when reasoning is enabled")
					assert.NotEmpty(t, reasoningText, "Reasoning text should not be empty")
					t.Logf("Reasoning: %s", reasoningText)
				} else if !tt.enabled {
					assert.False(t, hasReasoning, "Expected no reasoning part when reasoning is disabled")
				}
			}

			assert.NotEmpty(t, finalAnswer, "Final answer should not be empty")
			t.Logf("Final answer: %s", finalAnswer)

			// Verify usage includes reasoning tokens if applicable
			if hasReasoning && resp.Usage != nil {
				assert.Greater(t, resp.Usage.TotalTokens, 0, "Total tokens should be greater than 0")
				t.Logf("Token usage - Input: %d, Output: %d, Total: %d",
					resp.Usage.InputTokens, resp.Usage.OutputTokens, resp.Usage.TotalTokens)
			}
		})
	}
}

// TestLive_GenerateWithReasoningStreaming tests the reasoning feature with streaming responses.
// Note: Currently, the Anthropic API may not stream reasoning content in real-time;
// the reasoning part might be included in the final accumulated message.
//
// To run these reasoning tests:
// ANTHROPIC_API_KEY=your_key go test -v -run TestLive_GenerateWithReasoning
func TestLive_GenerateWithReasoningStreaming(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	ctx := context.Background()
	g, err := genkit.Init(ctx, genkit.WithPlugins(&anthropic.Plugin{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}))
	require.NoError(t, err)

	// Use claude-3.7-sonnet which supports reasoning
	model := anthropic.Model(g, "claude-3.7-sonnet")
	require.NotNil(t, model)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	config := struct {
		ai.GenerationCommonConfig
		anthropic.ExtendedThinkingConfig
	}{
		GenerationCommonConfig: ai.GenerationCommonConfig{
			MaxOutputTokens: 5000, // Large enough to support reasoning budget (5000 * 0.25 = 1250 > 1024)
			Temperature:     0.0,
		},
		ExtendedThinkingConfig: anthropic.ExtendedThinkingConfig{
			ExtendedThinkingEnabled:     true,
			ExtendedThinkingBudgetRatio: 0.25, // 25% of maxOutputTokens
		},
	}

	req := &ai.ModelRequest{
		Messages: []*ai.Message{
			{
				Role: ai.RoleUser,
				Content: []*ai.Part{
					ai.NewTextPart("What is 25% of 80? Show your reasoning step by step."),
				},
			},
		},
		Config: config,
	}

	var chunks []string
	var chunkTypes []string
	resp, err := model.Generate(ctx, req, func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
		if len(chunk.Content) > 0 {
			for _, part := range chunk.Content {
				if part.IsText() && part.Text != "" {
					chunks = append(chunks, part.Text)
					chunkTypes = append(chunkTypes, "text")
				} else if part.IsReasoning() && part.Text != "" {
					chunks = append(chunks, part.Text)
					chunkTypes = append(chunkTypes, "reasoning")
				}
			}
		}
		return nil
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, chunks)

	// Check final response structure
	hasReasoning := false
	var reasoningText string
	var finalAnswer string

	for _, part := range resp.Message.Content {
		if part.IsReasoning() {
			hasReasoning = true
			reasoningText = part.Text
		} else if part.IsText() {
			finalAnswer = part.Text
		}
	}

	assert.True(t, hasReasoning, "Expected reasoning part in final response")
	assert.NotEmpty(t, reasoningText, "Reasoning text should not be empty")
	assert.NotEmpty(t, finalAnswer, "Final answer should not be empty")

	// Log streaming details
	t.Logf("Received %d chunks during streaming", len(chunks))
	t.Logf("Chunk types: %v", chunkTypes)
	t.Logf("Final reasoning: %s", reasoningText)
	t.Logf("Final answer: %s", finalAnswer)
}

// TestLive_Claude4AutomaticReasoning tests Claude 4's automatic reasoning capabilities.
//
// With the new defaultModelParams implementation:
// - Claude 4 models have reasoning enabled by default (Enabled: true, BudgetRatio: 0)
// - When budgetRatio is 0, it defaults to 25% of maxOutputTokens
// - Users can override this by explicitly setting ExtendedThinkingConfig
//
// This test verifies both the default automatic reasoning and explicit configuration overrides.
func TestLive_Claude4AutomaticReasoning(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	ctx := context.Background()
	g, err := genkit.Init(ctx, genkit.WithPlugins(&anthropic.Plugin{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}))
	require.NoError(t, err)

	tests := []struct {
		name            string
		modelName       string
		prompt          string
		expectReasoning bool
	}{
		{
			name:            "Claude 4 - Simple calculation",
			modelName:       "claude-4-sonnet",
			prompt:          "What is 2+2?",
			expectReasoning: true, // Claude 4 has reasoning enabled by default
		},
		{
			name:            "Claude 4 - Complex problem",
			modelName:       "claude-4-sonnet",
			prompt:          "A snail climbs up a wall. During the day, it climbs 3 feet up, but at night it slides 2 feet down. If the wall is 10 feet high, how many days will it take the snail to reach the top? Think through this step by step.",
			expectReasoning: true,
		},
		{
			name:            "Claude 3.7 - Logic puzzle",
			modelName:       "claude-3.7-sonnet",
			prompt:          "Three friends (Alice, Bob, and Carol) are sitting in a row. Alice is not on the left. Bob is not on the right. Carol is not in the middle. What is the order from left to right?",
			expectReasoning: true, // Claude 3.7 also has reasoning enabled by default
		},
		{
			name:            "Claude 3.5 - Simple problem",
			modelName:       "claude-3.5-haiku",
			prompt:          "If a train travels 120 miles in 2 hours, what is its speed in mph? Think step by step.",
			expectReasoning: false, // Claude 3.5 has reasoning disabled by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := anthropic.Model(g, tt.modelName)
			if model == nil {
				t.Skipf("Model %s not available", tt.modelName)
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
			defer cancel()

			// Test without providing ExtendedThinkingConfig at all
			// This should use the defaultModelParams
			req := &ai.ModelRequest{
				Messages: []*ai.Message{
					{
						Role: ai.RoleUser,
						Content: []*ai.Part{
							ai.NewTextPart(tt.prompt),
						},
					},
				},
				Config: ai.GenerationCommonConfig{
					MaxOutputTokens: 5000, // Ensure enough tokens for reasoning (5000 * 0.25 = 1250 > 1024)
					Temperature:     0.0,
				},
			}

			resp, err := model.Generate(ctx, req, nil)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Message)
			require.NotEmpty(t, resp.Message.Content)

			// Check if reasoning was automatically used
			hasReasoning := false
			var reasoningText string
			var finalAnswer string

			for _, part := range resp.Message.Content {
				if part.IsReasoning() {
					hasReasoning = true
					reasoningText = part.Text
					t.Logf("Found reasoning part with %d characters", len(reasoningText))
				} else if part.IsText() {
					finalAnswer = part.Text
				}
			}

			t.Logf("Model %s - Has reasoning: %v", tt.modelName, hasReasoning)
			if hasReasoning {
				t.Logf("Reasoning text preview: %.200s...", reasoningText)
			}

			// Verify based on expected behavior from defaultModelParams
			assert.Equal(t, tt.expectReasoning, hasReasoning,
				"Model %s should have reasoning=%v based on defaultModelParams",
				tt.modelName, tt.expectReasoning)

			assert.NotEmpty(t, finalAnswer, "Final answer should not be empty")
			t.Logf("Final answer: %s", finalAnswer)

			// Verify usage includes reasoning tokens if applicable
			if hasReasoning && resp.Usage != nil {
				assert.Greater(t, resp.Usage.TotalTokens, 0, "Total tokens should be greater than 0")
				t.Logf("Token usage - Input: %d, Output: %d, Total: %d",
					resp.Usage.InputTokens, resp.Usage.OutputTokens, resp.Usage.TotalTokens)
			}
		})
	}
}

// TestLive_Claude4ExplicitReasoningControl tests that users can override the default reasoning behavior
// by explicitly setting ExtendedThinkingConfig
func TestLive_Claude4ExplicitReasoningControl(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	ctx := context.Background()
	g, err := genkit.Init(ctx, genkit.WithPlugins(&anthropic.Plugin{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}))
	require.NoError(t, err)

	tests := []struct {
		name             string
		modelName        string
		prompt           string
		extendedThinking *bool // nil means not set, true/false means explicitly set
		budgetRatio      float64
		expectReasoning  bool
	}{
		{
			name:             "Claude 4 - Explicitly disable reasoning",
			modelName:        "claude-4-sonnet",
			prompt:           "A complex math problem: If x + 2y = 10 and 3x - y = 5, what are x and y?",
			extendedThinking: ptr(false),
			budgetRatio:      0,
			expectReasoning:  false, // Should not have reasoning when explicitly disabled
		},
		{
			name:             "Claude 3.5 - Explicitly enable reasoning (silently ignored)",
			modelName:        "claude-3.5-haiku",
			prompt:           "What is 10 * 5? Think step by step.",
			extendedThinking: ptr(true),
			budgetRatio:      0.25,
			expectReasoning:  false, // Claude 3.5 doesn't support reasoning, API silently ignores it
		},
		{
			name:             "Claude 3.7 - Override default with disable",
			modelName:        "claude-3.7-sonnet",
			prompt:           "Solve: 2x + 3 = 7",
			extendedThinking: ptr(false),
			budgetRatio:      0,
			expectReasoning:  false, // Should not have reasoning when explicitly disabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := anthropic.Model(g, tt.modelName)
			if model == nil {
				t.Skipf("Model %s not available", tt.modelName)
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
			defer cancel()

			// Build config based on test case
			var config interface{}
			if tt.extendedThinking != nil {
				// Explicitly set ExtendedThinkingConfig
				config = map[string]interface{}{
					"maxOutputTokens":             2000,
					"temperature":                 0.0,
					"extendedThinkingEnabled":     *tt.extendedThinking,
					"extendedThinkingBudgetRatio": tt.budgetRatio,
				}
			} else {
				// Only set GenerationCommonConfig
				config = ai.GenerationCommonConfig{
					MaxOutputTokens: 2000,
					Temperature:     0.0,
				}
			}

			req := &ai.ModelRequest{
				Messages: []*ai.Message{
					{
						Role: ai.RoleUser,
						Content: []*ai.Part{
							ai.NewTextPart(tt.prompt),
						},
					},
				},
				Config: config,
			}

			resp, err := model.Generate(ctx, req, nil)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Message)
			require.NotEmpty(t, resp.Message.Content)

			// Check if reasoning was used
			hasReasoning := false
			var reasoningText string
			var finalAnswer string

			for _, part := range resp.Message.Content {
				if part.IsReasoning() {
					hasReasoning = true
					reasoningText = part.Text
					t.Logf("Found reasoning part with %d characters", len(reasoningText))
				} else if part.IsText() {
					finalAnswer = part.Text
				}
			}

			t.Logf("Model %s with explicit config - Has reasoning: %v", tt.modelName, hasReasoning)
			if hasReasoning {
				t.Logf("Reasoning text preview: %.200s...", reasoningText)
			}

			// Verify based on expected behavior
			if tt.modelName == "claude-3.5-haiku" && tt.expectReasoning {
				// Claude 3.5 doesn't support reasoning, so skip assertion
				t.Logf("Note: Claude 3.5 doesn't support reasoning feature")
			} else {
				assert.Equal(t, tt.expectReasoning, hasReasoning,
					"Model %s should have reasoning=%v based on explicit config",
					tt.modelName, tt.expectReasoning)
			}

			assert.NotEmpty(t, finalAnswer, "Final answer should not be empty")
			t.Logf("Final answer: %s", finalAnswer)
		})
	}
}

// Helper function to create a pointer to a bool
func ptr(b bool) *bool {
	return &b
}
