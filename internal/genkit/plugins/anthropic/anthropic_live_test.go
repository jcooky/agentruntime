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
			name:      "claude-3.5-sonnet simple",
			modelName: "claude-3.5-sonnet",
			prompt:    "What is 2+2? Answer with just the number.",
			timeout:   30 * time.Second,
		},
		{
			name:      "claude-3.7-sonnet simple",
			modelName: "claude-3.7-sonnet",
			prompt:    "What is the capital of Japan? Answer with just the city name.",
			timeout:   30 * time.Second,
		},
		{
			name:      "claude-4-sonnet simple",
			modelName: "claude-4-sonnet",
			prompt:    "What is the capital of France? Answer with just the city name.",
			timeout:   30 * time.Second,
		},
		{
			name:      "claude-3.5-sonnet with reasoning",
			modelName: "claude-3.5-sonnet",
			prompt:    "If a train travels 120 miles in 2 hours, what is its speed in mph? Answer with just the number.",
			timeout:   30 * time.Second,
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

	ctx := context.Background()
	g, err := genkit.Init(ctx, genkit.WithPlugins(&anthropic.Plugin{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}))
	require.NoError(t, err)

	model := anthropic.Model(g, "claude-3.5-sonnet")
	require.NotNil(t, model)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	model := anthropic.Model(g, "claude-3.5-sonnet")
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

	ctx := context.Background()
	g, err := genkit.Init(ctx, genkit.WithPlugins(&anthropic.Plugin{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}))
	require.NoError(t, err)

	model := anthropic.Model(g, "claude-3.5-sonnet")
	require.NotNil(t, model)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
