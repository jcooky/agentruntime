package config

import (
	goconfig "github.com/golobby/config/v3"
	"github.com/golobby/config/v3/pkg/feeder"
	"github.com/pkg/errors"
	"os"
)

func resolveConfig[T any](config *T, testing bool) error {
	if config == nil {
		return errors.New("config is nil")
	}

	configReader := goconfig.New()
	if err := configReader.Feed(); err != nil {
		return errors.Wrapf(err, "failed to load config")
	}

	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		configReader.AddFeeder(feeder.DotEnv{Path: ".env"})
	}

	filename := ".env.test"
	if v := os.Getenv("ENV_TEST_FILE"); v != "" {
		filename = v
	}
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		configReader.AddFeeder(feeder.DotEnv{Path: filename})
	}

	configReader.AddFeeder(feeder.Env{})

	return errors.Wrapf(configReader.AddStruct(config).Feed(), "failed to load config")
}
