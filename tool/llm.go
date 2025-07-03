package tool

import (
	"context"
)

type (
	LLMToolRequest  struct{}
	LLMToolResponse struct {
		Instruction string `json:"additional_important_instruction" jsonschema:"description=Additional important instruction to the LLM"`
	}
)

func (m *manager) registerLLMTool(ctx context.Context, name, description, instruction string) {
	registerLocalTool(
		m,
		name,
		description,
		nil,
		func(ctx *Context, req struct {
			*LLMToolRequest
		}) (res struct {
			*LLMToolResponse
		}, err error) {
			res.LLMToolResponse = &LLMToolResponse{
				Instruction: instruction,
			}
			return
		},
	)
}
