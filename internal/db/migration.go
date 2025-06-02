package db

import (
	"context"
	"fmt"

	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/errors"
	"gorm.io/gorm"
)

var (
	schema = "agentnetwork"
)

func AutoMigrate(ctx context.Context, db *gorm.DB) error {
	if err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)).Error; err != nil {
		return errors.Wrapf(err, "failed to create schema")
	}

	_, tx := OpenSession(ctx, db)

	return errors.WithStack(tx.AutoMigrate(
		&entity.Message{},
		&entity.Thread{},
		&entity.AgentRuntime{},
		&entity.Mention{},
	))
}

func DropAll(ctx context.Context, db *gorm.DB) error {
	_, tx := OpenSession(ctx, db)
	return errors.WithStack(tx.Migrator().DropTable(
		&entity.Mention{},
		&entity.AgentRuntime{},
		&entity.Thread{},
		&entity.Message{},
	))
}
