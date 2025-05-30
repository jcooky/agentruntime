package db

import (
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/errors"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	if err := db.Exec("CREATE SCHEMA IF NOT EXISTS agentruntime").Error; err != nil {
		return errors.Wrapf(err, "failed to create schema")
	}

	return errors.WithStack(db.AutoMigrate(
		&entity.Message{},
		&entity.Thread{},
		&entity.AgentRuntime{},
		&entity.Mention{},
	))
}

func DropAll(db *gorm.DB) error {
	return errors.WithStack(db.Migrator().DropTable(
		"thread_participants",
		&entity.Mention{},
		&entity.AgentRuntime{},
		&entity.Thread{},
		&entity.Message{},
	))
}
