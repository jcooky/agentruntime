package config

import (
	"context"

	"github.com/habiliai/agentruntime/internal/di"
)

type LogConfig struct {
	LogLevel   string `env:"LOG_LEVEL"`
	LogHandler string `env:"LOG_HANDLER"`
}

var LogConfigKey = di.NewKey()

func init() {
	di.Register(LogConfigKey, func(ctx context.Context, c *di.Container) (any, error) {
		config := LogConfig{
			LogLevel:   "debug",
			LogHandler: "default",
		}
		return &config, resolveConfig(&config, c.Env == di.EnvTest)
	})
}
