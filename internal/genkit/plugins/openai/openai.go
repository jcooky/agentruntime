package openai

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/internal/config"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/internal/openaiapi"
	goopenai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"os"
)

const (
	provider    = "openai"
	labelPrefix = "OpenAI"
	apiKeyEnv   = "OPENAI_API_KEY"
)

var (
	knownCaps = map[string]ai.ModelSupports{
		"o3":                          config.BasicText,
		"o4-mini":                     config.BasicText,
		goopenai.ChatModelO3Mini:      config.BasicText,
		goopenai.ChatModelO1:          config.BasicText,
		goopenai.ChatModelGPT4o:       config.Multimodal,
		goopenai.ChatModelGPT4oMini:   config.Multimodal,
		goopenai.ChatModelGPT4Turbo:   config.Multimodal,
		goopenai.ChatModelGPT4:        config.BasicText,
		goopenai.ChatModelGPT3_5Turbo: config.BasicText,
	}

	knownEmbedders = []string{
		goopenai.EmbeddingModelTextEmbedding3Small,
		goopenai.EmbeddingModelTextEmbedding3Large,
		goopenai.EmbeddingModelTextEmbeddingAda002,
	}
)

type Plugin struct {
	// The API key to access the service.
	// If empty, the values of the environment variables OPENAI_API_KEY will be consulted.
	APIKey string
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
			return fmt.Errorf("OpenAI requires setting %s in the environment. You can get an API key at https://platform.openai.com/api-keys", apiKeyEnv)
		}
	}

	client := goopenai.NewClient(
		option.WithAPIKey(apiKey),
	)

	for model, caps := range knownCaps {
		openaiapi.DefineModel(g, &client, labelPrefix, provider, model, caps)
	}

	for _, e := range knownEmbedders {
		openaiapi.DefineEmbedder(g, &client, provider, e)
	}

	return nil
}

// Model returns the [ai.Model] with the given name.
// It returns nil if the model was not defined.
func Model(g *genkit.Genkit, name string) ai.Model {
	return genkit.LookupModel(g, provider, name)
}

// Embedder returns the [ai.Embedder] with the given name.
// It returns nil if the embedder was not defined.
func Embedder(g *genkit.Genkit, name string) ai.Embedder {
	return genkit.LookupEmbedder(g, provider, name)
}
