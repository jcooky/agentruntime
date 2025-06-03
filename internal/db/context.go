package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	sessionCtxKey = uuid.New()
)

func OpenSession(ctx context.Context, db *gorm.DB) (context.Context, *gorm.DB) {
	tx, ok := ctx.Value(sessionCtxKey).(*gorm.DB)
	if ok {
		return ctx, tx
	}

	return WithSession(ctx, db)
}

func WithSession(ctx context.Context, db *gorm.DB) (context.Context, *gorm.DB) {
	tx := db.WithContext(ctx)

	// Only execute schema commands for PostgreSQL databases
	if db.Dialector.Name() == "postgres" {
		tx.Exec(fmt.Sprintf("SET search_path TO %s", schema))
	}

	return context.WithValue(ctx, sessionCtxKey, tx), tx
}
