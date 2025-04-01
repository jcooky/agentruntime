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

func LoadAgentFromFile(file string) (agent AgentConfig, err error) {
	var yamlBytes []byte
	if yamlBytes, err = os.ReadFile(file); err != nil {
		err = errors.Wrapf(err, "failed to read file %s", file)
		return
	}

	if err = yaml.Unmarshal(yamlBytes, &agent); err != nil {
		err = errors.Wrapf(err, "failed to unmarshal file %s", file)
		return
	}

	return
}

func LoadAgentsFromFiles(files []string) ([]AgentConfig, error) {
	agents := make([]AgentConfig, 0, len(files))
	for _, file := range files {
		agent, err := LoadAgentFromFile(file)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	return agents, nil
}
