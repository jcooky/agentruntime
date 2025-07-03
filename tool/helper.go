package tool

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/entity"
)

func registerLocalTool[In any, Out any](m *manager, name, description string, skill *entity.NativeAgentSkill, fn func(ctx *Context, input In) (Out, error)) ai.Tool {
	return genkit.DefineTool(
		m.genkit,
		name,
		description,
		func(ctx *ai.ToolContext, input In) (Out, error) {
			out, err := fn(&Context{
				Context: ctx,
				skill:   skill,
			}, input)
			if err == nil {
				appendCallData(ctx, CallData{
					Name:      name,
					Arguments: input,
					Result:    out,
				})
			}
			return out, err
		},
	)
}
