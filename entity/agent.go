package entity

type Agent struct {
	AgentCard
	ModelName       string             `json:"modelName,omitempty"`
	ModelConfig     map[string]any     `json:"modelConfig,omitempty"`
	System          string             `json:"system,omitempty"`
	Role            string             `json:"role,omitempty"`
	Instruction     string             `json:"instruction,omitempty"`
	MessageExamples [][]MessageExample `json:"messageExamples,omitempty"`
	Knowledge       []map[string]any   `json:"knowledge,omitempty"`
	Evaluator       AgentEvaluator     `json:"evaluator,omitempty"`

	Metadata map[string]string `json:"metadata"`
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
