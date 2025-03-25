package tool

import (
	"context"
	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/db"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"sync"
)

type (
	manager struct {
		logger   *mylog.Logger
		db       *gorm.DB
		config   *config.RuntimeConfig
		mcpTools map[string]ai.Tool

		mtx     sync.Mutex
		closeFn []func()
	}
)

func (m *manager) GetLocalTool(_ context.Context, toolName string) ai.Tool {
	return ai.LookupTool(toolName)
}

func (m *manager) GetTools(ctx context.Context, names []string) ([]entity.Tool, error) {
	_, tx := db.OpenSession(ctx, m.db)

	var tools []entity.Tool
	if err := tx.Find(&tools, "name IN ?", names).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find tools")
	}

	return tools, nil
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
