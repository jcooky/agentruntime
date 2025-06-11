package memory

import (
	"context"

	"github.com/pkg/errors"
)

func (s *SqliteService) SetContext(ctx context.Context, context *AgentContext) error {
	tx := s.db.WithContext(ctx)

	if context == nil {
		return errors.New("context cannot be nil")
	}

	if err := tx.Save(context).Error; err != nil {
		return errors.Wrapf(err, "failed to save cursor for %s", context.Name)
	}

	return nil
}

func (s *SqliteService) GetContext(ctx context.Context, name string) (*AgentContext, error) {
	tx := s.db.WithContext(ctx)

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
