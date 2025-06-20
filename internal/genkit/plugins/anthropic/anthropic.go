package anthropic

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/internal/config"
)

const (
	provider    = "anthropic"
	labelPrefix = "Anthropic"
	apiKeyEnv   = "ANTHROPIC_API_KEY"
)

var (
	knownCaps = map[string]ai.ModelSupports{
		"claude-opus-4-20250514":   config.Multimodal,
		"claude-sonnet-4-20250514": config.Multimodal,
		"claude-3-7-sonnet-latest": config.Multimodal,
		"claude-3-5-sonnet-latest": config.Multimodal,
		"claude-3-opus-latest":     config.Multimodal,
	}
	defaultRequestTimeout = 10 * time.Minute
)

type Plugin struct {
	// The API key to access the service for Anthropic.
	// If empty, the values of the environment variables ANTHROPIC_API_KEY will be consulted.
	APIKey string

	// The timeout for requests to the Anthropic API.
	// If empty, the default timeout of 10 minutes will be used.
	RequestTimeout time.Duration
}

var (
	_ genkit.Plugin = (*Plugin)(nil)
)

// Name implements genkit.Plugin.
func (o *Plugin) Name() string {
	return provider
}

// Init implements genkit.Plugin.
// After calling Init, you may call [DefineModel] to create and register any additional generative models.
func (o *Plugin) Init(_ context.Context, g *genkit.Genkit) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s.Init: %w", provider, err)
		}
	}()

	apiKey := o.APIKey
	if apiKey == "" {
		apiKey = os.Getenv(apiKeyEnv)
		if apiKey == "" {
			return fmt.Errorf("Anthropic API key not found in environment variable: %s", apiKeyEnv)
		}
	}

	if o.RequestTimeout == 0 {
		o.RequestTimeout = defaultRequestTimeout
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
		option.WithRequestTimeout(o.RequestTimeout),
	)

	// Define models with simplified names as requested
	DefineModel(g, &client, labelPrefix, provider, "claude-4-opus", "claude-opus-4-20250514", knownCaps["claude-opus-4-20250514"])
	DefineModel(g, &client, labelPrefix, provider, "claude-4-sonnet", "claude-sonnet-4-20250514", knownCaps["claude-sonnet-4-20250514"])

	// Also define Claude 3.7 and 3.5 models as alternatives
	DefineModel(g, &client, labelPrefix, provider, "claude-3.7-sonnet", "claude-3-7-sonnet-latest", knownCaps["claude-3-7-sonnet-latest"])
	DefineModel(g, &client, labelPrefix, provider, "claude-3.5-sonnet", "claude-3-5-sonnet-latest", knownCaps["claude-3-5-sonnet-latest"])
	DefineModel(g, &client, labelPrefix, provider, "claude-3-opus", "claude-3-opus-latest", knownCaps["claude-3-opus-latest"])

	return nil
}

// Model returns the [ai.Model] with the given name.
// It returns nil if the model was not defined.
func Model(g *genkit.Genkit, name string) ai.Model {
	return genkit.LookupModel(g, provider, name)
}
