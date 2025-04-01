package tool

import (
	"context"
	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/mylog"
	"gorm.io/gorm"
	"strings"
)

type (
	manager struct {
		logger *mylog.Logger
		db     *gorm.DB
		config *config.RuntimeConfig

		closeFn []func()
	}
)

func (m *manager) GetTool(_ context.Context, toolName string) ai.Tool {
	toolName = strings.Replace(toolName, "/", "_", -1)
	tool := ai.LookupTool(toolName)
	if tool.Action() == nil {
		return nil
	}

	return tool
}

func (m *manager) Close() {
	for _, closeFn := range m.closeFn {
		closeFn()
	}
}

var (
	_ Manager = (*manager)(nil)
)

func RegisterLocalTool[In any, Out any](name string, description string, fn func(context.Context, In) (Out, error)) ai.Tool {
	return ai.DefineTool(
		name,
		description,
		func(ctx context.Context, input In) (Out, error) {
			out, err := fn(ctx, input)
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
