package config

import (
	"github.com/jcooky/go-din"
)

type NetworkConfig struct {
	LogConfig
	Host                string `env:"HOST"`
	Port                int    `env:"PORT"`
	DatabaseUrl         string `env:"DATABASE_URL"`
	DatabaseAutoMigrate bool   `env:"DATABASE_AUTO_MIGRATE"`
}

func init() {
	din.RegisterT(func(c *din.Container) (*NetworkConfig, error) {
		conf := &NetworkConfig{
			LogConfig: LogConfig{
				LogLevel:   "debug",
				LogHandler: "default",
			},
			Host:                "0.0.0.0",
			Port:                9080,
			DatabaseUrl:         "postgres://postgres:postgres@localhost:5432/test?search_path=agentruntime",
			DatabaseAutoMigrate: true,
		}

		return conf, resolveConfig(
			conf,
			c.Env == din.EnvTest,
		)
	})
}
