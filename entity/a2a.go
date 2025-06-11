package entity

// AgentProvider represents the service provider of an agent.
type AgentProvider struct {
	// Agent provider's organization name.
	Organization string `json:"organization"`
	// Agent provider's URL.
	URL string `json:"url"`
}

// AgentCard conveys key information about an agent:
// - Overall details (version, name, description, uses)
// - Skills: A set of capabilities the agent can perform
// - Default modalities/content types supported by the agent.
// - Authentication requirements
type AgentCard struct {
	// Human readable name of the agent.
	// Example: "Recipe Agent"
	Name string `json:"name"`
	// A human-readable description of the agent. Used to assist users and
	// other agents in understanding what the agent can do.
	// Example: "Agent that helps users with recipes and cooking."
	Description string `json:"description"`
	// A URL to the address the agent is hosted at.
	URL string `json:"url"`
	// A URL to an icon for the agent.
	IconURL *string `json:"iconUrl,omitempty"`
	// The service provider of the agent.
	Provider *AgentProvider `json:"provider,omitempty"`
	// The version of the agent - format is up to the provider.
	// Example: "1.0.0"
	Version string `json:"version"`
	// A URL to documentation for the agent.
	DocumentationURL *string `json:"documentationUrl,omitempty"`
	// The set of interaction modes that the agent supports across all skills.
	// This can be overridden per-skill.
	// Supported media types for input.
	DefaultInputModes []string `json:"defaultInputModes"`
	// Supported media types for output.
	DefaultOutputModes []string `json:"defaultOutputModes"`
}
