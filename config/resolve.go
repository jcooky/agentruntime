package config

import (
	"os"

	goconfig "github.com/golobby/config/v3"
	"github.com/golobby/config/v3/pkg/feeder"
	"github.com/habiliai/agentruntime/errors"
)

func resolveConfig[T any](config *T, testing bool) error {
	if config == nil {
		return errors.New("config is nil")
	}

	configReader := goconfig.New()
	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		configReader.AddFeeder(feeder.DotEnv{Path: ".env"})
	}

	filename := os.Getenv("ENV_TEST_FILE")
	if filename != "" {
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			configReader.AddFeeder(feeder.DotEnv{Path: filename})
		}
	}

	if err := configReader.
		AddFeeder(feeder.Env{}).
		AddStruct(config).
		Feed(); err != nil {
		return errors.Wrapf(err, "failed to load config")
	}

	return nil
}
