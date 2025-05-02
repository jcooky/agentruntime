package config

import "github.com/firebase/genkit/go/ai"

type GenerationReasoningConfig struct {
	ai.GenerationCommonConfig
	ReasoningEffort string `json:"reasoningEffort,omitempty"`
}
