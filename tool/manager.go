package tool

import (
	"context"
	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/db"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"gorm.io/gorm"
)

type (
	Manager interface {
		GetTool(ctx context.Context, toolName string) ai.Tool
		GetMCPTools(ctx context.Context, serverName string) []ai.Tool
		RegisterMCPTool(ctx context.Context, req RegisterMCPToolRequest) error
	}
)

var (
	ManagerKey = di.NewKey()
)

func init() {
	di.Register(ManagerKey, func(ctx context.Context, env di.Env) (any, error) {
		conf, err := di.Get[*config.RuntimeConfig](ctx, config.RuntimeConfigKey)
		if err != nil {
			return nil, err
		}

		s := &manager{
			db:     di.MustGet[*gorm.DB](ctx, db.Key),
			logger: di.MustGet[*mylog.Logger](ctx, mylog.Key),
			config: conf,
		}

		go func() {
			<-ctx.Done()
			s.Close()
		}()

		return s, nil
	})
}
