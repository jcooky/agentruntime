package runtime

import (
	"context"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/db"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/tool"
	"github.com/pkg/errors"
	"github.com/yukinagae/genkit-go-plugins/plugins/openai"
	"gorm.io/gorm"
	"os"
)

type (
	Runtime interface {
		Run(ctx context.Context, threadIds uint, agentIds []uint) error
	}
	service struct {
		logger      *mylog.Logger
		db          *gorm.DB
		toolManager tool.Manager
	}
)

var (
	Key = di.NewKey()
)

func init() {
	di.Register(Key, func(c context.Context, _ di.Env) (any, error) {
		conf, err := di.Get[*config.RuntimeConfig](c, config.RuntimeConfigKey)
		if err != nil {
			return nil, err
		}

		os.Setenv("OPENAI_API_KEY", conf.OpenAIApiKey)
		if err := openai.Init(c, &openai.Config{
			APIKey: conf.OpenAIApiKey,
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		return &service{
			logger:      di.MustGet[*mylog.Logger](c, mylog.Key),
			db:          di.MustGet[*gorm.DB](c, db.Key),
			toolManager: di.MustGet[tool.Manager](c, tool.ManagerKey),
		}, nil
	})
}
