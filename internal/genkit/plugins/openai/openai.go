package openai

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/core"
	"os"

	"github.com/firebase/genkit/go/genkit"

	"github.com/firebase/genkit/go/ai"

	goopenai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	provider    = "openai"
	labelPrefix = "OpenAI"
	apiKeyEnv   = "OPENAI_API_KEY"
)

var (
	knownCaps = map[string]ai.ModelSupports{
		"o3":                          BasicText,
		"o4-mini":                     BasicText,
		goopenai.ChatModelO3Mini:      BasicText,
		goopenai.ChatModelO1:          BasicText,
		goopenai.ChatModelGPT4o:       Multimodal,
		goopenai.ChatModelGPT4oMini:   Multimodal,
		goopenai.ChatModelGPT4Turbo:   Multimodal,
		goopenai.ChatModelGPT4:        BasicText,
		goopenai.ChatModelGPT3_5Turbo: BasicText,
	}

	modelsSupportingResponseFormats = []string{
		goopenai.ChatModelO3Mini,
		goopenai.ChatModelO1,
		goopenai.ChatModelGPT4o,
		goopenai.ChatModelGPT4oMini,
		goopenai.ChatModelGPT4Turbo,
		goopenai.ChatModelGPT3_5Turbo,
		"o3",
		"o4-mini",
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
		defineModel(g, &client, model, caps)
	}

	for _, e := range knownEmbedders {
		defineEmbedder(g, &client, e)
	}

	return nil
}

// requires state.mu
func defineModel(g *genkit.Genkit, client *goopenai.Client, name string, caps ai.ModelSupports) ai.Model {
	meta := &ai.ModelInfo{
		Label:    labelPrefix + " - " + name,
		Supports: &caps,
	}
	return genkit.DefineModel(
		g,
		provider,
		name,
		meta,
		func(ctx context.Context, req *ai.ModelRequest, _ core.StreamCallback[*ai.ModelResponseChunk]) (*ai.ModelResponse, error) {
			return generate(ctx, client, name, req)
		},
	)
}

// requires state.mu
func defineEmbedder(g *genkit.Genkit, client *goopenai.Client, name string) ai.Embedder {
	return genkit.DefineEmbedder(g, provider, name, func(ctx context.Context, input *ai.EmbedRequest) (*ai.EmbedResponse, error) {
		var data goopenai.EmbeddingNewParamsInputUnion
		for _, doc := range input.Input {
			for _, p := range doc.Content {
				data.OfArrayOfStrings = append(data.OfArrayOfStrings, p.Text)
			}
		}

		params := goopenai.EmbeddingNewParams{
			Input:          data,
			Model:          name,
			EncodingFormat: goopenai.EmbeddingNewParamsEncodingFormatFloat,
		}

		embRes, err := client.Embeddings.New(ctx, params)
		if err != nil {
			return nil, err
		}

		var res ai.EmbedResponse
		for _, emb := range embRes.Data {
			embedding := make([]float32, len(emb.Embedding))
			for i, val := range emb.Embedding {
				embedding[i] = float32(val)
			}
			res.Embeddings = append(res.Embeddings, &ai.Embedding{Embedding: embedding})
		}
		return &res, nil
	})
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

func generate(
	ctx context.Context,
	client *goopenai.Client,
	model string,
	input *ai.ModelRequest,
) (*ai.ModelResponse, error) {
	req, err := convertRequest(model, input)
	if err != nil {
		return nil, err
	}

	res, err := client.Chat.Completions.New(ctx, req)
	if err != nil {
		return nil, err
	}

	jsonMode := false
	if input.Output != nil &&
		input.Output.Format == ai.OutputFormatJSON {
		jsonMode = true
	}

	r := translateResponse(res, jsonMode)
	r.Request = input
	return r, nil
}
