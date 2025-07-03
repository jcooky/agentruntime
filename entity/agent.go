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
	Skills []AgentSkillUnion `json:"skills"`

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
