package config

type (
	ModelConfig struct {
		OpenAIAPIKey    string `json:"openaiApiKey"`
		XAIAPIKey       string `json:"xaiApiKey"`
		AnthropicAPIKey string `json:"anthropicApiKey"`
		TraceVerbose    bool   `json:"traceVerbose"`
	}
)
