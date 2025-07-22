package config

type MemoryConfig struct {
	GenerationModel string `json:"generationModel"`
}

func NewMemoryConfig() *MemoryConfig {
	return &MemoryConfig{
		GenerationModel: "openai/o4-mini",
	}
}
