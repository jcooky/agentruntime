package config

import (
	"fmt"
	"os"
)

type MemoryConfig struct {
	SqliteEnabled bool   `env:"SQLITE_ENABLED"`
	SqlitePath    string `env:"SQLITE_PATH"`
	VectorEnabled bool   `env:"VECTOR_ENABLED"`
}

func NewMemoryConfig() *MemoryConfig {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return &MemoryConfig{
		SqliteEnabled: true,
		SqlitePath:    fmt.Sprintf("%s/.agentruntime/memory.db", home),
		VectorEnabled: true,
	}
}
