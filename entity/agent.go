package entity

type Agent struct {
	Name            string             `json:"name,omitempty"`
	ModelName       string             `json:"model_name,omitempty"`
	ModelConfig     map[string]any     `json:"model_config,omitempty"`
	System          string             `json:"system,omitempty"`
	Role            string             `json:"role,omitempty"`
	Bio             []string           `json:"bio,omitempty"`
	Lore            []string           `json:"lore,omitempty"`
	MessageExamples [][]MessageExample `json:"message_examples,omitempty"`
	Knowledge       []map[string]any   `json:"knowledge,omitempty"`
	Evaluator       AgentEvaluator     `json:"evaluator,omitempty"`

	Tools []Tool `json:"tools"`

	Metadata map[string]string `json:"metadata"`
}

type Tool struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type MessageExample struct {
	User    string   `json:"user,omitempty"`
	Text    string   `json:"text,omitempty"`
	Actions []string `json:"actions,omitempty"`
}

type AgentEvaluator struct {
	Prompt     string `json:"prompt,omitempty"`
	NumRetries int    `json:"num_retries,omitempty"`
}
