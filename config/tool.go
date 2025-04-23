package config

import (
	"context"
	"github.com/habiliai/agentruntime/internal/di"
)

type ToolConfig struct {
	OpenWeatherApiKey string `env:"OPENWEATHER_API_KEY"`
	SerpApiKey        string `env:"SERP_API_KEY"`
}

var ToolConfigKey = di.NewKey()

func init() {
	di.Register(ToolConfigKey, func(ctx context.Context, env di.Env) (any, error) {
		c := ToolConfig{
			OpenWeatherApiKey: "",
			SerpApiKey:        "",
		}

		if err := resolveConfig(&c, env == di.EnvTest); err != nil {
			return nil, err
		}

		return &c, nil
	})
}
