package runtime

import (
	"context"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/stringslices"
	"github.com/habiliai/agentruntime/thread"
	"github.com/habiliai/agentruntime/tool"
	"github.com/pkg/errors"
	"github.com/yukinagae/genkit-go-plugins/plugins/openai"
	"os"
	"strings"
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
	di.Register(ServiceKey, func(c context.Context, _ di.Env) (any, error) {
		conf, err := di.Get[*config.RuntimeConfig](c, config.RuntimeConfigKey)
		if err != nil {
			return nil, err
		}

		os.Setenv("OPENAI_API_KEY", conf.OpenAIApiKey)
		if err := openai.Init(c, &openai.Config{
			APIKey: conf.OpenAIApiKey,
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		logger := di.MustGet[*mylog.Logger](c, mylog.Key)

		return &service{
			logger:              logger,
			toolManager:         di.MustGet[tool.Manager](c, tool.ManagerKey),
			threadManagerClient: di.MustGet[thread.ThreadManagerClient](c, thread.ClientKey),
		}, nil
	})
}
