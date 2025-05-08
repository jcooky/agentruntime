package config

import (
	"github.com/jcooky/go-din"
)

type LogConfig struct {
	LogLevel     string `env:"LOG_LEVEL"`
	LogHandler   string `env:"LOG_HANDLER"`
	TraceVerbose bool   `env:"TRACE_VERBOSE"`
}

func init() {
	din.RegisterT(func(c *din.Container) (*LogConfig, error) {
		config := LogConfig{
			LogLevel:   "debug",
			LogHandler: "default",
		}
		return &config, resolveConfig(&config, c.Env == din.EnvTest)
	})
}
