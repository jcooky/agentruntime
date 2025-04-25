package config

import (
	"context"

	"github.com/habiliai/agentruntime/internal/di"
)

type NetworkConfig struct {
	LogConfig
	Host                string `env:"HOST"`
	Port                int    `env:"PORT"`
	DatabaseUrl         string `env:"DATABASE_URL"`
	DatabaseAutoMigrate bool   `env:"DATABASE_AUTO_MIGRATE"`
}

var (
	NetworkConfigKey = di.NewKey()
)

func resolveNetworkConfig(testing bool) (*NetworkConfig, error) {
	c := NetworkConfig{
		LogConfig: LogConfig{
			LogLevel:   "debug",
			LogHandler: "default",
		},
		Host:                "0.0.0.0",
		Port:                9080,
		DatabaseUrl:         "postgres://postgres:postgres@localhost:5432/test?search_path=agentruntime",
		DatabaseAutoMigrate: true,
	}

	if err := resolveConfig(&c, testing); err != nil {
		return nil, err
	}

	return &c, nil
}

func init() {
	di.Register(NetworkConfigKey, func(ctx context.Context, c *di.Container) (any, error) {
		return resolveNetworkConfig(
			c.Env == di.EnvTest,
		)
	})
}
