package memory

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/jcooky/go-din"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type (
	Service interface {
		SetContext(ctx context.Context, context *AgentContext) error
		GetContext(ctx context.Context, name string) (*AgentContext, error)
	}
	SqliteService struct {
		db *gorm.DB
	}
	AgentContext struct {
		Name       string `gorm:"primaryKey"`
		LastCursor uint   `gorm:"not null"`
	}
)

var (
	_                 Service = (*SqliteService)(nil)
	SqliteServiceName         = din.NewRandomName()
)

func init() {
	din.Register(SqliteServiceName, func(c *din.Container) (any, error) {
		logger := din.MustGet[*slog.Logger](c, mylog.Key)
		conf := din.MustGetT[*config.MemoryConfig](c)

		if !conf.SqliteEnabled {
			return nil, errors.New("sqlite memory service is not enabled. Please check your configuration.")
		}
		if conf.SqlitePath == "" {
			return nil, errors.New("sqlite memory service path is not configured. Please check your configuration.")
		} else if _, err := os.Stat(conf.SqlitePath); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(conf.SqlitePath), 0755); err != nil {
				return nil, errors.Wrapf(err, "failed to create sqlite directory at %s", conf.SqlitePath)
			} else {
				logger.Info("created sqlite directory", slog.String("path", conf.SqlitePath))
			}
		}
		db, err := gorm.Open(
			sqlite.Open(
				fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_foreign_keys=on", conf.SqlitePath),
			),
			&gorm.Config{},
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open sqlite database at %s", conf.SqlitePath)
		}

		if err := db.AutoMigrate(&AgentContext{}); err != nil {
			return nil, errors.Wrapf(err, "failed to auto-migrate sqlite database at %s", conf.SqlitePath)
		}
		c.RegisterOnShutdown(func(_ context.Context) {
			db, err := db.DB()
			if err != nil {
				logger.Warn("failed to get database connection", slog.Any("error", err))
			}
			if err := db.Close(); err != nil {
				logger.Warn("failed to close database connection", slog.Any("error", err))
			}
		})

		return &SqliteService{db: db}, nil
	})
}
