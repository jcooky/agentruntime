package entity

type Agent struct {
	AgentCard       `json:",inline"`
	ModelName       string             `json:"model,omitempty"`
	ModelConfig     map[string]any     `json:"modelConfig,omitempty"`
	System          string             `json:"system,omitempty"`
	Role            string             `json:"role,omitempty"`
	Prompt          string             `json:"prompt,omitempty"`
	MessageExamples [][]MessageExample `json:"messageExamples,omitempty"`
	Knowledge       []map[string]any   `json:"knowledge,omitempty"`
	Evaluator       AgentEvaluator     `json:"evaluator,omitempty"`

	// Skills are a unit of capability that an agent can perform.
	Skills []AgentSkill `json:"skills"`

	Metadata map[string]any `json:"metadata"`
}

type MessageExample struct {
	User    string   `json:"user,omitempty"`
	Text    string   `json:"text,omitempty"`
	Actions []string `json:"actions,omitempty"`
}

type AgentEvaluator struct {
	Prompt     string `json:"prompt,omitempty"`
	NumRetries int    `json:"numRetries,omitempty"`
}

// AgentSkill represents a unit of capability that an agent can perform.
type AgentSkill struct {
	ID   string `json:"id" jsonschema:"required,description=Field for unique identify to skill"`
	Type string `json:"type" jsonschema:"required,enum=llm,enum=mcp,enum=nativeTool"`

	Tools   []string `json:"tools,omitempty" jsonschema_description:"MCP tools name"`
	Command string   `json:"command,omitempty" jsonschema_description:"Command to run MCP server"`
	Args    []string `json:"args,omitempty" jsonschema_description:"Arguments to run MCP server"`

	// Remote MCP Support
	URL       string                 `json:"url,omitempty" jsonschema_description:"URL for remote MCP server (SSE, OAuth-SSE, or Streamable)"`
	Transport string                 `json:"transport,omitempty" jsonschema_description:"Transport type: stdio, sse, oauth-sse, http. Auto-detected if not specified"`
	Headers   map[string]string      `json:"headers,omitempty" jsonschema_description:"HTTP headers for authentication (e.g., API keys)"`
	OAuth     *AgentSkillOAuthConfig `json:"oauth,omitempty" jsonschema_description:"OAuth configuration for oauth-sse transport"`

	Env map[string]any `json:"env,omitempty" jsonschema_description:"It can be environment variables for MCP or can be configuration for nativeTool"`

	Name        string `json:"name,omitempty" jsonschema_description:"name for LLM tool or native tool. It can be also mcp server name"`
	Description string `json:"description,omitempty" jsonschema_description:"It uses only when type is nativeTool or llm. Use default description owned tool if empty and type is nativeTool"`
	Instruction string `json:"instruction,omitempty" jsonschema_description:"It uses only when type is llm."`
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
