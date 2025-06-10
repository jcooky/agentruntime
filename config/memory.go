package config

import (
	"fmt"
	"os"
)

type MemoryConfig struct {
	SqliteEnabled bool   `json:"sqliteEnabled"`
	SqlitePath    string `json:"sqlitePath"`
}

func NewMemoryConfig() *MemoryConfig {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return &MemoryConfig{
		SqliteEnabled: true,
		SqlitePath:    fmt.Sprintf("%s/.agentruntime/memory.db", home),
	}
}
