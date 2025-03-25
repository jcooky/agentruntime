package config

import (
	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
	"os"
)

type AgentConfig struct {
	Name            string   `yaml:"name"`
	System          string   `yaml:"system"`
	Role            string   `yaml:"role"`
	Bio             []string `yaml:"bio"`
	Lore            []string `yaml:"lore"`
	MessageExamples []struct {
		Messages []struct {
			Name    string   `yaml:"name"`
			Text    string   `yaml:"text"`
			Actions []string `yaml:"actions"`
		} `yaml:"messages"`
	} `yaml:"messageExamples"`
	Model      string               `yaml:"model"`
	Tools      []string             `yaml:"tools"`
	Metadata   map[string]string    `yaml:"metadata"`
	Knowledge  []map[string]any     `yaml:"knowledge"`
	MCPServers map[string]MCPServer `yaml:"mcpServers"`
}

func LoadAgentsFromFiles(files []string) ([]AgentConfig, error) {
	agents := make([]AgentConfig, 0, len(files))
	for _, file := range files {
		var agent AgentConfig
		yamlBytes, err := os.ReadFile(file)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read file %s", file)
		}

		if err := yaml.Unmarshal(yamlBytes, &agent); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal file %s", file)
		}
		agents = append(agents, agent)
	}
	return agents, nil
}
