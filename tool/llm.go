package tool

import (
	"context"

	"github.com/habiliai/agentruntime/entity"
	"github.com/pkg/errors"
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

func (m *manager) registerLLMSkill(ctx context.Context, skill *entity.LLMAgentSkill) error {
	if skill.Name == "" {
		return errors.New("llm name is required")
	}
	if skill.Description == "" {
		return errors.New("llm description is required")
	}
	if skill.Instruction == "" {
		return errors.New("llm instruction is required")
	}
	m.registerLLMTool(ctx, skill.Name, skill.Description, skill.Instruction)

	return nil
}
