package anthropic

type ExtendedThinkingConfig struct {
	ExtendedThinkingEnabled     bool    `json:"extendedThinkingEnabled,omitempty"`
	ExtendedThinkingBudgetRatio float64 `json:"extendedThinkingBudgetRatio,omitempty"`
}
