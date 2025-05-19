package memory

import (
	"context"

	"github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/internal/db"
)

func (s *SqliteService) SetContext(ctx context.Context, context *AgentContext) error {
	_, tx := db.OpenSession(ctx, s.db)

	if context == nil {
		return errors.New("context cannot be nil")
	}

	if err := tx.Save(context).Error; err != nil {
		return errors.Wrapf(err, "failed to save cursor for %s", context.Name)
	}

	return nil
}

func (s *SqliteService) GetContext(ctx context.Context, name string) (*AgentContext, error) {
	_, tx := db.OpenSession(ctx, s.db)

	var agentContext AgentContext
	if r := tx.Find(&agentContext, "name = ?", name); r.Error != nil {
		return nil, errors.Wrapf(r.Error, "failed to get cursor for %s", name)
	} else if r.RowsAffected == 0 {
		// No record found, return a new AgentContext with LastCursor set to 0
		return &AgentContext{
			Name:       name,
			LastCursor: 0,
		}, nil
	}

	return &agentContext, nil
}
