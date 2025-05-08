package config

import (
	"github.com/jcooky/go-din"
)

type (
	OpenAIConfig struct {
		APIKey string `env:"OPENAI_API_KEY"`
	}
	XAIConfig struct {
		APIKey string `env:"XAI_API_KEY"`
	}
)

func init() {
	din.RegisterT(func(c *din.Container) (*OpenAIConfig, error) {
		conf := &OpenAIConfig{}
		return conf, resolveConfig(conf, c.Env == din.EnvTest)
	})
	din.RegisterT(func(c *din.Container) (*XAIConfig, error) {
		conf := &XAIConfig{}
		return conf, resolveConfig(conf, c.Env == din.EnvTest)
	})
}
