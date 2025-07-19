package memory_test

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKey_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live test in short mode")
	}

	// Check API keys - need at least one [[memory:3743077]]
	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if openaiKey == "" && anthropicKey == "" {
		t.Skip("Skipping live test: OPENAI_API_KEY or ANTHROPIC_API_KEY required. Run with: godotenv go test ./memory -v")
	}

	ctx := t.Context()
	store := memory.NewInMemoryStore()

	service, err := memory.NewServiceWithStore(ctx, store, &config.ModelConfig{
		OpenAIAPIKey:    openaiKey,
		AnthropicAPIKey: anthropicKey,
	}, slog.Default())
	require.NoError(t, err)

	tests := []struct {
		name         string
		input        string
		tags         []string
		existingKeys []string
		expectPrefix string
	}{
		{
			name:         "user personal info",
			input:        "My name is John Smith",
			tags:         []string{"personal"},
			existingKeys: []string{},
			expectPrefix: "user_",
		},
		{
			name:         "user preference",
			input:        "I like dark roast coffee with oat milk",
			tags:         []string{"personal", "preferences"},
			existingKeys: []string{},
			expectPrefix: "user_preference_",
		},
		{
			name:         "project decision",
			input:        "We decided to use React for the frontend",
			tags:         []string{"work", "decisions"},
			existingKeys: []string{},
			expectPrefix: "decision_",
		},
		{
			name:         "avoid duplication",
			input:        "I live in Seoul, South Korea",
			tags:         []string{"personal"},
			existingKeys: []string{"user_name_full", "user_preference_coffee", "user_location_city"},
			expectPrefix: "user_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := service.GenerateKey(ctx, tt.input, tt.tags, "", tt.existingKeys)
			require.NoError(t, err)

			// Basic validations
			assert.NotEmpty(t, key, "Generated key should not be empty")
			assert.True(t, strings.HasPrefix(key, tt.expectPrefix), "Key should start with expected prefix: %s, got: %s", tt.expectPrefix, key)
			assert.True(t, strings.Contains(key, "_"), "Key should contain underscores")
			assert.Equal(t, strings.ToLower(key), key, "Key should be lowercase")

			// Check no duplication with existing keys
			for _, existingKey := range tt.existingKeys {
				assert.NotEqual(t, existingKey, key, "Generated key should not duplicate existing key: %s", existingKey)
			}

			t.Logf("Input: %s → Generated key: %s", tt.input, key)
		})
	}
}

func TestGenerateTags_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live test in short mode")
	}

	// Check API keys [[memory:3743077]]
	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if openaiKey == "" && anthropicKey == "" {
		t.Skip("Skipping live test: OPENAI_API_KEY or ANTHROPIC_API_KEY required. Run with: godotenv go test ./memory -v")
	}

	ctx := t.Context()
	store := memory.NewInMemoryStore()

	service, err := memory.NewServiceWithStore(ctx, store, &config.ModelConfig{
		OpenAIAPIKey:    openaiKey,
		AnthropicAPIKey: anthropicKey,
	}, slog.Default())
	require.NoError(t, err)

	tests := []struct {
		name          string
		input         string
		existingTags  []string
		expectContain []string
	}{
		{
			name:          "personal preference",
			input:         "I love drinking dark roast coffee in the morning",
			existingTags:  []string{},
			expectContain: []string{"personal", "preferences"},
		},
		{
			name:          "work decision",
			input:         "Our team chose to implement microservices architecture",
			existingTags:  []string{},
			expectContain: []string{"work", "decisions"},
		},
		{
			name:          "learning goal",
			input:         "I want to learn Python programming this year",
			existingTags:  []string{},
			expectContain: []string{"goals", "skills"},
		},
		{
			name:          "reuse existing tags",
			input:         "I also like green tea",
			existingTags:  []string{"personal", "preferences", "work", "goals"},
			expectContain: []string{"personal", "preferences"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, err := service.GenerateTags(ctx, tt.input, "", tt.existingTags)
			require.NoError(t, err)

			// Basic validations
			assert.NotEmpty(t, tags, "Generated tags should not be empty")
			assert.True(t, len(tags) >= 1 && len(tags) <= 3, "Should generate 1-3 tags, got: %d", len(tags))

			// Check all tags are lowercase
			for _, tag := range tags {
				assert.Equal(t, strings.ToLower(tag), tag, "Tag should be lowercase: %s", tag)
			}

			// Check if contains expected tags (at least one)
			hasExpected := false
			for _, expectedTag := range tt.expectContain {
				for _, generatedTag := range tags {
					if generatedTag == expectedTag {
						hasExpected = true
						break
					}
				}
				if hasExpected {
					break
				}
			}
			assert.True(t, hasExpected, "Generated tags %v should contain at least one of expected tags %v", tags, tt.expectContain)

			t.Logf("Input: %s → Generated tags: %v", tt.input, tags)
		})
	}
}

func TestGenerateKeyAndTags_Integration_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live test in short mode")
	}

	// Check API keys [[memory:3743077]]
	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if openaiKey == "" && anthropicKey == "" {
		t.Skip("Skipping live test: OPENAI_API_KEY or ANTHROPIC_API_KEY required. Run with: godotenv go test ./memory -v")
	}

	ctx := t.Context()
	store := memory.NewInMemoryStore()

	service, err := memory.NewServiceWithStore(ctx, store, &config.ModelConfig{
		OpenAIAPIKey:    openaiKey,
		AnthropicAPIKey: anthropicKey,
	}, slog.Default())
	require.NoError(t, err)

	// Simulate building up memory over multiple interactions
	inputs := []string{
		"My name is Alice Johnson",
		"I work as a software engineer at Google",
		"I love drinking espresso",
		"Our team decided to use TypeScript for the new project",
		"I want to learn machine learning this year",
	}

	var allKeys []string
	var allTags []string

	for i, input := range inputs {
		t.Run(fmt.Sprintf("interaction_%d", i+1), func(t *testing.T) {
			// Generate tags first
			tags, err := service.GenerateTags(ctx, input, "", allTags)
			require.NoError(t, err)

			// Generate key with context from tags
			key, err := service.GenerateKey(ctx, input, tags, "", allKeys)
			require.NoError(t, err)

			// Validate key format
			assert.True(t, strings.Contains(key, "_"), "Key should contain underscores")
			assert.Equal(t, strings.ToLower(key), key, "Key should be lowercase")

			// Validate tags
			assert.True(t, len(tags) >= 1 && len(tags) <= 3, "Should have 1-3 tags")

			// Check uniqueness
			for _, existingKey := range allKeys {
				assert.NotEqual(t, existingKey, key, "Key should be unique")
			}

			// Add to collections for next iteration
			allKeys = append(allKeys, key)
			for _, tag := range tags {
				if !contains(allTags, tag) {
					allTags = append(allTags, tag)
				}
			}

			t.Logf("Input: %s", input)
			t.Logf("  → Key: %s", key)
			t.Logf("  → Tags: %v", tags)
			t.Logf("  → All keys so far: %v", allKeys)
			t.Logf("  → All tags so far: %v", allTags)
		})
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
