package entity_test

import (
	"encoding/json"
	"testing"

	"github.com/habiliai/agentruntime/entity"
	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshalSkill(t *testing.T) {
	tests := []struct {
		name     string
		skill    *entity.AgentSkillUnion
		expected string
	}{
		{
			name: "MCP Skill",
			skill: &entity.AgentSkillUnion{
				Type: entity.AgentSkillTypeMCP,
				OfMCP: &entity.MCPAgentSkill{
					ID:      "mcp-skill",
					Name:    "test-mcp",
					Command: "test-command",
					Args:    []string{"arg1", "arg2"},
				},
			},
			expected: `{"args":["arg1","arg2"],"command":"test-command","id":"mcp-skill","name":"test-mcp","type":"mcp"}`,
		},
		{
			name: "LLM Skill",
			skill: &entity.AgentSkillUnion{
				Type: entity.AgentSkillTypeLLM,
				OfLLM: &entity.LLMAgentSkill{
					ID:          "llm-skill",
					Name:        "test-llm",
					Description: "test description",
					Instruction: "test instruction",
				},
			},
			expected: `{"description":"test description","id":"llm-skill","instruction":"test instruction","name":"test-llm","type":"llm"}`,
		},
		{
			name: "Native Skill",
			skill: &entity.AgentSkillUnion{
				Type: entity.AgentSkillTypeNative,
				OfNative: &entity.NativeAgentSkill{
					ID:      "native-skill",
					Name:    "test-native",
					Details: "test details",
				},
			},
			expected: `{"details":"test details","id":"native-skill","name":"test-native","type":"nativeTool"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			jsonData, err := json.Marshal(tt.skill)
			require.NoError(t, err)
			require.JSONEq(t, tt.expected, string(jsonData))

			// Test unmarshaling
			var unmarshaled entity.AgentSkillUnion
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			require.Equal(t, tt.skill.Type, unmarshaled.Type)

			// Test that the round trip works
			jsonData2, err := json.Marshal(&unmarshaled)
			require.NoError(t, err)
			require.JSONEq(t, tt.expected, string(jsonData2))
		})
	}
}
