package config

import (
	"fmt"
	"os"

	"github.com/habiliai/agentruntime/errors"
	"github.com/jcooky/go-din"
)

type MemoryConfig struct {
	SqliteEnabled bool   `env:"SQLITE_ENABLED"`
	SqlitePath    string `env:"SQLITE_PATH"`
	VectorEnabled bool   `env:"VECTOR_ENABLED"`
}

func init() {
	din.RegisterT(func(c *din.Container) (*MemoryConfig, error) {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get user home directory")
		}
		conf := &MemoryConfig{
			SqliteEnabled: true,
			SqlitePath:    fmt.Sprintf("%s/.agentruntime/memory.db", home),
			VectorEnabled: true,
		}

		return conf, resolveConfig(
			conf,
			c.Env == din.EnvTest,
		)
	})
}
