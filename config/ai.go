package config

import (
	"context"

	"github.com/habiliai/agentruntime/internal/di"
)

type (
	OpenAIConfig struct {
		APIKey string `env:"OPENAI_API_KEY"`
	}
	XAIConfig struct {
		APIKey string `env:"XAI_API_KEY"`
	}
)

var (
	OpenAIConfigKey = di.NewKey()
	XAIConfigKey    = di.NewKey()
)

func init() {
	di.Register(OpenAIConfigKey, func(ctx context.Context, c *di.Container) (any, error) {
		conf := &OpenAIConfig{}
		return conf, resolveConfig(conf, c.Env == di.EnvTest)
	})
	di.Register(XAIConfigKey, func(ctx context.Context, c *di.Container) (any, error) {
		conf := &XAIConfig{}
		return conf, resolveConfig(conf, c.Env == di.EnvTest)
	})
}
