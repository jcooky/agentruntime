package config

import (
	"context"

	"github.com/habiliai/agentruntime/internal/di"
)

type RuntimeConfig struct {
	LogConfig
	OpenAIConfig
	ToolConfig
	Host              string `env:"HOST"`
	Port              int    `env:"PORT"`
	NetworkGrpcAddr   string `env:"NETWORK_GRPC_ADDR"`
	NetworkGrpcSecure bool   `env:"NETWORK_GRPC_SECURE"`
	RuntimeGrpcAddr   string `env:"RUNTIME_GRPC_ADDR"`
}

var (
	RuntimeConfigKey = di.NewKey()
)

func resolveRuntimeConfig(testing bool) (*RuntimeConfig, error) {
	c := RuntimeConfig{
		LogConfig: LogConfig{
			LogLevel:   "debug",
			LogHandler: "default",
		},
		Host:              "0.0.0.0",
		Port:              10080,
		NetworkGrpcAddr:   "127.0.0.1:9080",
		NetworkGrpcSecure: false,
		RuntimeGrpcAddr:   "127.0.0.1:10080",
	}

	if err := resolveConfig(&c, testing); err != nil {
		return nil, err
	}

	return &c, nil
}

func init() {
	di.Register(RuntimeConfigKey, func(ctx context.Context, c *di.Container) (any, error) {
		return resolveRuntimeConfig(c.Env == di.EnvTest)
	})
}
