package config

import (
	"context"
	goconfig "github.com/golobby/config/v3"
	"github.com/golobby/config/v3/pkg/feeder"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/pkg/errors"
	"os"
)

type RuntimeConfig struct {
	Host              string `env:"HOST"`
	Port              int    `env:"PORT"`
	LogLevel          string `env:"LOG_LEVEL"`
	LogHandler        string `env:"LOG_HANDLER"`
	OpenAIApiKey      string `env:"OPENAI_API_KEY"`
	OpenWeatherApiKey string `env:"OPENWEATHER_API_KEY"`
	NetworkGrpcAddr   string `env:"NETWORK_GRPC_ADDR"`
	NetworkGrpcSecure bool   `env:"NETWORK_GRPC_SECURE"`
}

var (
	RuntimeConfigKey = di.NewKey()
)

func resolveRuntimeConfig(testing bool) (*RuntimeConfig, error) {
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

	c := RuntimeConfig{
		Host:              "0.0.0.0",
		Port:              10080,
		LogLevel:          "debug",
		LogHandler:        "default",
		NetworkGrpcAddr:   "localhost:9080",
		NetworkGrpcSecure: false,
	}
	if err := configReader.AddStruct(&c).Feed(); err != nil {
		return nil, err
	}

	return &c, nil
}

func init() {
	di.Register(RuntimeConfigKey, func(ctx context.Context, env di.Env) (any, error) {
		return resolveRuntimeConfig(env == di.EnvTest)
	})
}
