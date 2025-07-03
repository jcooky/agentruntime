package anthropic

import (
	"context"
	"testing"

	"github.com/firebase/genkit/go/genkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlugin_Init(t *testing.T) {
	tests := []struct {
		name    string
		plugin  *Plugin
		envKey  string
		wantErr bool
	}{
		{
			name: "with API key",
			plugin: &Plugin{
				APIKey: "test-api-key",
			},
			wantErr: false,
		},
		{
			name:    "with env API key",
			plugin:  &Plugin{},
			envKey:  "test-env-key",
			wantErr: false,
		},
		{
			name:    "no API key",
			plugin:  &Plugin{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envKey != "" {
				t.Setenv("ANTHROPIC_API_KEY", tt.envKey)
			} else if tt.name == "no API key" {
				// Explicitly unset the API key for this test
				t.Setenv("ANTHROPIC_API_KEY", "")
			}

			ctx := context.Background()
			_, err := genkit.Init(ctx, genkit.WithPlugins(tt.plugin))

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlugin_Name(t *testing.T) {
	plugin := &Plugin{}
	assert.Equal(t, "anthropic", plugin.Name())
}

func TestModel(t *testing.T) {
	ctx := context.Background()
	g, err := genkit.Init(ctx, genkit.WithPlugins(&Plugin{
		APIKey: "test-key",
	}))
	require.NoError(t, err)

	tests := []struct {
		name      string
		modelName string
		wantNil   bool
	}{
		{
			name:      "claude-4-opus exists",
			modelName: "claude-4-opus",
			wantNil:   false,
		},
		{
			name:      "claude-4-sonnet exists",
			modelName: "claude-4-sonnet",
			wantNil:   false,
		},
		{
			name:      "claude-3.7-sonnet exists",
			modelName: "claude-3.7-sonnet",
			wantNil:   false,
		},
		{
			name:      "unknown model",
			modelName: "claude-2-legacy",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := Model(g, tt.modelName)
			if tt.wantNil {
				assert.Nil(t, model)
			} else {
				assert.NotNil(t, model)
			}
		})
	}
}

func TestKnownModels(t *testing.T) {
	ctx := context.Background()
	g, err := genkit.Init(ctx, genkit.WithPlugins(&Plugin{
		APIKey: "test-key",
	}))
	require.NoError(t, err)

	// Test that known models are registered with correct capabilities
	opus := Model(g, "claude-4-opus")
	require.NotNil(t, opus)

	sonnet := Model(g, "claude-4-sonnet")
	require.NotNil(t, sonnet)

	sonnet37 := Model(g, "claude-3.7-sonnet")
	require.NotNil(t, sonnet37)

	haiku := Model(g, "claude-3.5-haiku")
	require.NotNil(t, haiku)
}
