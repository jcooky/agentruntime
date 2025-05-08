package config

import (
	"github.com/jcooky/go-din"
)

type ToolConfig struct {
	OpenWeatherApiKey string `env:"OPENWEATHER_API_KEY"`
	SerpApiKey        string `env:"SERP_API_KEY"`
}

func init() {
	din.RegisterT(func(c *din.Container) (*ToolConfig, error) {
		conf := &ToolConfig{
			OpenWeatherApiKey: "",
			SerpApiKey:        "",
		}

		if err := resolveConfig(conf, c.Env == din.EnvTest); err != nil {
			return nil, err
		}

		return conf, nil
	})
}
