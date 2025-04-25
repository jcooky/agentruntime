package runtime

import (
	"context"
	"strings"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/stringslices"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/thread"
	"github.com/habiliai/agentruntime/tool"
	"github.com/pkg/errors"
)

type (
	Service interface {
		RegisterAgent(
			ctx context.Context,
			ac config.AgentConfig,
		) (*entity.Agent, error)
		Run(ctx context.Context, threadIds uint, agents []entity.Agent) error
		findAgentsByNames(names []string) ([]entity.Agent, error)
	}
	service struct {
		logger              *mylog.Logger
		toolManager         tool.Manager
		agents              []entity.Agent
		threadManagerClient thread.ThreadManagerClient
		networkClient       network.AgentNetworkClient
		runner              engine.Engine
	}
)

var (
	_          Service = (*service)(nil)
	ServiceKey         = di.NewKey()
)

func (s *service) findAgentsByNames(names []string) ([]entity.Agent, error) {
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
	di.Register(ServiceKey, func(c context.Context, container *di.Container) (any, error) {
		logger := di.MustGet[*mylog.Logger](c, container, mylog.Key)

		return &service{
			logger:              logger,
			runner:              di.MustGet[engine.Engine](c, container, engine.Key),
			toolManager:         di.MustGet[tool.Manager](c, container, tool.ManagerKey),
			threadManagerClient: di.MustGet[thread.ThreadManagerClient](c, container, thread.ClientKey),
			networkClient:       di.MustGet[network.AgentNetworkClient](c, container, network.ClientKey),
		}, nil
	})
}
