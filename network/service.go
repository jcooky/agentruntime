package network

import (
	"context"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/db"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/stringslices"
	"github.com/habiliai/agentruntime/tool"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log/slog"
	"strings"
	"time"
)

type (
	Service interface {
		GetAgentRuntimeInfo(ctx context.Context, agentNames []string) ([]entity.AgentRuntime, error)
		GetAllAgentRuntimeInfo(ctx context.Context) ([]entity.AgentRuntime, error)
		RegisterAgent(ctx context.Context, addr string, secure bool, agentInfo []*AgentInfo) error
		DeregisterAgent(ctx context.Context, agentNames []string) error
		CheckLive(ctx context.Context, agentNames []string) error
	}

	service struct {
		logger      *slog.Logger
		db          *gorm.DB
		toolManager tool.Manager
	}
)

var (
	_ Service = (*service)(nil)
)

func (s *service) CheckLive(ctx context.Context, agentNames []string) error {
	_, tx := db.OpenSession(ctx, s.db)
	for _, agentName := range agentNames {
		var agentRuntime entity.AgentRuntime
		if err := s.db.First(&agentRuntime, "name = ?", agentName).Error; err != nil {
			return errors.Wrapf(err, "failed to find agent runtime")
		}

		agentRuntime.LastLiveAt = time.Now()

		if err := agentRuntime.Save(tx); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) DeregisterAgent(ctx context.Context, agentNames []string) error {
	_, tx := db.OpenSession(ctx, s.db)
	return tx.Transaction(func(tx *gorm.DB) error {
		for _, agentName := range agentNames {
			var agentRuntime entity.AgentRuntime
			if err := tx.First(&agentRuntime, "name = ?", agentName).Error; err != nil {
				return errors.Wrapf(err, "failed to find agent runtime")
			}

			if err := agentRuntime.Delete(tx); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *service) RegisterAgent(ctx context.Context, addr string, secure bool, agentInfo []*AgentInfo) error {
	_, tx := db.OpenSession(ctx, s.db)

	return tx.Transaction(func(tx *gorm.DB) error {
		for _, info := range agentInfo {
			var agentRuntime entity.AgentRuntime
			if err := tx.Clauses(clause.Locking{
				Strength: "UPDATE",
			}).Find(&agentRuntime, "lower(name) = ?", strings.ToLower(info.Name)).Error; err != nil {
				return errors.Wrapf(err, "failed to find agent runtime")
			}

			agentRuntime.Name = info.Name
			agentRuntime.Addr = addr
			agentRuntime.Secure = secure
			agentRuntime.Role = info.Role
			agentRuntime.Metadata = datatypes.NewJSONType(info.Metadata)
			agentRuntime.LastLiveAt = time.Now()

			if err := agentRuntime.Save(tx); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *service) GetAgentRuntimeInfo(ctx context.Context, agentNames []string) ([]entity.AgentRuntime, error) {
	_, tx := db.OpenSession(ctx, s.db)

	agentNames = stringslices.ToLower(agentNames)

	var agentRuntimes []entity.AgentRuntime
	if err := tx.Where("lower(name) IN ?", agentNames).Find(&agentRuntimes).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find agent runtime")
	}

	return agentRuntimes, nil
}

func (s *service) GetAllAgentRuntimeInfo(ctx context.Context) ([]entity.AgentRuntime, error) {
	_, tx := db.OpenSession(ctx, s.db)

	var agentRuntimes []entity.AgentRuntime
	if err := tx.Find(&agentRuntimes).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find agent runtime")
	}

	return agentRuntimes, nil
}

var (
	ManagerKey = di.NewKey()
)

func init() {
	di.Register(ManagerKey, func(c context.Context, _ di.Env) (any, error) {
		service := &service{
			logger:      di.MustGet[*slog.Logger](c, mylog.Key),
			db:          di.MustGet[*gorm.DB](c, db.Key),
			toolManager: di.MustGet[tool.Manager](c, tool.ManagerKey),
		}

		go service.runHealthChecker(c)

		return service, nil
	})
}
