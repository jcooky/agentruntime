package entity

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	AgentSkillTypeNative = "nativeTool"
	AgentSkillTypeLLM    = "llm"
	AgentSkillTypeMCP    = "mcp"
)

// AgentSkillUnion represents a unit of capability that an agent can perform.
type AgentSkillUnion struct {
	Type string `json:"type" jsonschema:"required,enum=llm,enum=mcp,enum=nativeTool"`

	OfMCP    *MCPAgentSkill    `json:",omitzero,inline"`
	OfLLM    *LLMAgentSkill    `json:",omitzero,inline"`
	OfNative *NativeAgentSkill `json:",omitzero,inline"`
}

type MCPAgentSkill struct {
	ID      string         `json:"id" jsonschema:"required,description=Field for unique identify to skill"`
	Name    string         `json:"name" jsonschema_description:"name for MCP server"`
	Tools   []string       `json:"tools,omitempty" jsonschema_description:"MCP tools name"`
	Command string         `json:"command" jsonschema_description:"Command to run MCP server"`
	Args    []string       `json:"args,omitempty" jsonschema_description:"Arguments to run MCP server"`
	Env     map[string]any `json:"env,omitempty" jsonschema_description:"It can be environment variables for MCP or can be configuration for nativeTool"`

	// Remote MCP Support
	URL       string                 `json:"url,omitempty" jsonschema_description:"URL for remote MCP server (SSE, OAuth-SSE, or Streamable)"`
	Transport string                 `json:"transport,omitempty" jsonschema_description:"Transport type: stdio, sse, oauth-sse, http. Auto-detected if not specified"`
	Headers   map[string]string      `json:"headers,omitempty" jsonschema_description:"HTTP headers for authentication (e.g., API keys)"`
	OAuth     *AgentSkillOAuthConfig `json:"oauth,omitempty" jsonschema_description:"OAuth configuration for oauth-sse transport"`
}

type LLMAgentSkill struct {
	ID          string `json:"id" jsonschema:"required,description=Field for unique identify to skill"`
	Name        string `json:"name" jsonschema_description:"name for LLM tool or native tool. It can be also mcp server name"`
	Description string `json:"description" jsonschema_description:"It uses only when type is nativeTool or llm. Use default description owned tool if empty and type is nativeTool"`
	Instruction string `json:"instruction" jsonschema_description:"It uses only when type is llm."`
}

type NativeAgentSkill struct {
	ID      string         `json:"id" jsonschema:"required,description=Field for unique identify to skill"`
	Name    string         `json:"name"`
	Details string         `json:"details"`
	Env     map[string]any `json:"env,omitempty" jsonschema_description:"It can be environment variables for MCP or can be configuration for nativeTool"`
}

// AgentSkillOAuthConfig represents OAuth configuration for AgentSkill
type AgentSkillOAuthConfig struct {
	ClientID              string   `json:"clientId,omitempty"`
	ClientSecret          string   `json:"clientSecret,omitempty"`
	AuthServerMetadataURL string   `json:"authServerMetadataUrl,omitempty"`
	RedirectURL           string   `json:"redirectUrl,omitempty"`
	Scopes                []string `json:"scopes,omitempty"`
	PKCEEnabled           bool     `json:"pkceEnabled,omitempty"`
}

func (u *AgentSkillUnion) UnmarshalJSON(data []byte) error {
	var tpe struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &tpe); err != nil {
		return errors.WithStack(err)
	}

	switch tpe.Type {
	case AgentSkillTypeMCP:
		u.Type = AgentSkillTypeMCP
		u.OfMCP = &MCPAgentSkill{}
		return errors.WithStack(json.Unmarshal(data, u.OfMCP))
	case AgentSkillTypeLLM:
		u.Type = AgentSkillTypeLLM
		u.OfLLM = &LLMAgentSkill{}
		return errors.WithStack(json.Unmarshal(data, u.OfLLM))
	case AgentSkillTypeNative:
		u.Type = AgentSkillTypeNative
		u.OfNative = &NativeAgentSkill{}
		return errors.WithStack(json.Unmarshal(data, u.OfNative))
	default:
		return errors.Errorf("unknown skill type: %s", tpe.Type)
	}
}
