package config

import (
	"github.com/jcooky/go-din"
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

func init() {
	din.RegisterT(func(c *din.Container) (*RuntimeConfig, error) {
		conf := &RuntimeConfig{
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

		if err := resolveConfig(&c, c.Env == din.EnvTest); err != nil {
			return nil, err
		}

		return conf, nil
	})
}
