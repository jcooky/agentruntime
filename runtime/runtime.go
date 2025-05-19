package runtime

import (
	"context"
	"log/slog"
	"strings"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/stringslices"
	"github.com/habiliai/agentruntime/internal/tool"
	"github.com/habiliai/agentruntime/network"
	"github.com/jcooky/go-din"
)

type (
	Service interface {
		RegisterAgent(
			ctx context.Context,
			ac config.AgentConfig,
		) (*entity.Agent, error)
		Run(ctx context.Context, threadIds uint, agents []entity.Agent) error
		FindAgentsByNames(names []string) ([]entity.Agent, error)
	}
	service struct {
		logger        *mylog.Logger
		toolManager   tool.Manager
		agents        []entity.Agent
		networkClient network.JsonRpcClient
		runner        engine.Engine
	}
)

var (
	_ Service = (*service)(nil)
)

func (s *service) FindAgentsByNames(names []string) ([]entity.Agent, error) {
	var (
		res   = make([]entity.Agent, 0, len(names))
		found = map[string]bool{}
	)
	for _, name := range names {
		found[strings.ToLower(name)] = false
	}
	for _, agent := range s.agents {
		if stringslices.ContainsIgnoreCase(names, agent.Name) {
			res = append(res, agent)
			found[strings.ToLower(agent.Name)] = true
		}
	}

	notFoundNames := make([]string, 0, len(names))
	for _, name := range names {
		if !found[strings.ToLower(name)] {
			notFoundNames = append(notFoundNames, name)
		}
	}
	if len(notFoundNames) > 0 {
		return nil, errors.Errorf("agent(s) %s not found", strings.Join(notFoundNames, ", "))
	}
	return res, nil
}

func init() {
	din.RegisterT(func(c *din.Container) (Service, error) {
		logger := din.MustGet[*slog.Logger](c, mylog.Key)

		return &service{
			logger:        logger,
			runner:        din.MustGetT[engine.Engine](c),
			toolManager:   din.MustGetT[tool.Manager](c),
			networkClient: din.MustGetT[network.JsonRpcClient](c),
		}, nil
	})
}
