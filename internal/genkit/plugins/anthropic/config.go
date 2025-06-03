package anthropic

type ExtendedThinkingConfig struct {
	Enabled      bool  `json:"enabled,omitempty"`
	BudgetTokens int64 `json:"budgetTokens,omitempty"`
}
