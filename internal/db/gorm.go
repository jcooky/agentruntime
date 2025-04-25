package db

import (
	"context"
	"time"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	Key = di.NewKey()
)

func OpenDB(databaseUrl string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseUrl))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return db, nil
}

func CloseDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return errors.Wrapf(err, "failed to get db")
	}
	if err := sqlDB.Close(); err != nil {
		return errors.Wrapf(err, "failed to close db")
	}

	return nil
}

func init() {
	di.Register(Key, func(ctx context.Context, c *di.Container) (any, error) {
		logger, err := di.Get[*mylog.Logger](ctx, c, mylog.Key)
		if err != nil {
			return nil, err
		}

		cfg, err := di.Get[*config.NetworkConfig](ctx, c, config.NetworkConfigKey)
		if err != nil {
			return nil, err
		}

		logger.Info("initialize database")
		db, err := OpenDB(cfg.DatabaseUrl)
		if err != nil {
			return nil, err
		}

		if c.Env == di.EnvTest {
			if err := DropAll(db); err != nil {
				return nil, errors.Wrapf(err, "failed to drop database")
			}
			time.Sleep(500 * time.Millisecond)
		}
		if cfg.DatabaseAutoMigrate || c.Env == di.EnvTest {
			if err := AutoMigrate(db); err != nil {
				return nil, errors.Wrapf(err, "failed to migrate database")
			}
		}

		go func() {
			<-ctx.Done()
			if err := CloseDB(db); err != nil {
				logger.Warn("failed to close database", "err", err)
			}
			logger.Info("database closed")
		}()

		return db, nil
	})
}
