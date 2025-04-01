package config

import (
	"context"
	goconfig "github.com/golobby/config/v3"
	"github.com/golobby/config/v3/pkg/feeder"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/pkg/errors"
	"os"
)

type NetworkConfig struct {
	Host                string `env:"HOST"`
	Port                int    `env:"PORT"`
	LogLevel            string `env:"LOG_LEVEL"`
	LogHandler          string `env:"LOG_HANDLER"`
	DatabaseUrl         string `env:"DATABASE_URL"`
	DatabaseAutoMigrate bool   `env:"DATABASE_AUTO_MIGRATE"`
}

var (
	NetworkConfigKey = di.NewKey()
)

func resolveNetworkConfig(testing bool) (*NetworkConfig, error) {
	configReader := goconfig.New()
	if err := configReader.Feed(); err != nil {
		return nil, errors.Wrapf(err, "failed to load config")
	}

	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		configReader.AddFeeder(feeder.DotEnv{Path: ".env"})
	}
	if testing {
		filename := ".env.test"
		if v := os.Getenv("ENV_TEST_FILE"); v != "" {
			filename = v
		}
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "should be existed %s in testing", filename)
		}
		configReader.AddFeeder(feeder.DotEnv{Path: filename})
	}
	configReader.AddFeeder(feeder.Env{})

	c := NetworkConfig{
		Host:                "0.0.0.0",
		Port:                9080,
		DatabaseUrl:         "postgres://postgres:postgres@localhost:5432/test?search_path=agentruntime",
		LogLevel:            "debug",
		LogHandler:          "default",
		DatabaseAutoMigrate: true,
	}
	if err := configReader.AddStruct(&c).Feed(); err != nil {
		return nil, err
	}

	return &c, nil
}

func init() {
	di.Register(NetworkConfigKey, func(ctx context.Context, env di.Env) (any, error) {
		return resolveNetworkConfig(env == di.EnvTest)
	})
}
