package config

import (
	"io"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/habiliai/agentruntime/errors"
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
	Model       string               `yaml:"model"`
	ModelConfig map[string]any       `yaml:"modelConfig"`
	Tools       []string             `yaml:"tools"`
	Metadata    map[string]string    `yaml:"metadata"`
	Knowledge   []map[string]any     `yaml:"knowledge"`
	MCPServers  map[string]MCPServer `yaml:"mcpServers"`
}

func LoadAgentFromFile(file io.Reader) (agent AgentConfig, err error) {
	var yamlBytes []byte
	if yamlBytes, err = io.ReadAll(file); err != nil {
		err = errors.Wrapf(err, "failed to read file")
		return
	}

	if err = yaml.Unmarshal(yamlBytes, &agent); err != nil {
		err = errors.Wrapf(err, "failed to unmarshal file")
		return
	}

	return
}

func LoadAgentsFromFiles(files []string) ([]AgentConfig, error) {
	agents := make([]AgentConfig, 0, len(files))
	for _, file := range files {
		fileReader, err := os.Open(file)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open file %s", file)
		}
		defer fileReader.Close()

		agent, err := LoadAgentFromFile(fileReader)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load agent from file %s", file)
		}
		agents = append(agents, agent)
	}
	return agents, nil
}
