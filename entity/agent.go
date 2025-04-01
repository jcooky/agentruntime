package entity

type Agent struct {
	Name            string
	ModelName       string
	System          string
	Role            string
	Bio             []string
	Lore            []string
	MessageExamples [][]MessageExample
	Knowledge       []map[string]any

	Tools []Tool

	Metadata map[string]string
}

type Tool struct {
	Name        string
	Description string
}

type MessageExample struct {
	User    string
	Text    string
	Actions []string
}
