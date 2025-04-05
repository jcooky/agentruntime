package entity

type Agent struct {
	Name            string             `json:"name"`
	ModelName       string             `json:"model_name"`
	System          string             `json:"system"`
	Role            string             `json:"role"`
	Bio             []string           `json:"bio"`
	Lore            []string           `json:"lore"`
	MessageExamples [][]MessageExample `json:"message_examples"`
	Knowledge       []map[string]any   `json:"knowledge"`

	Tools []Tool `json:"tools"`

	Metadata map[string]string `json:"metadata"`
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MessageExample struct {
	User    string   `json:"user"`
	Text    string   `json:"text"`
	Actions []string `json:"actions"`
}
