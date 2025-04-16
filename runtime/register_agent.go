package runtime

import (
	"context"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/entity"
)

func (s *service) RegisterAgent(
	ctx context.Context,
	ac config.AgentConfig,
) (*entity.Agent, error) {
	a, err := s.runner.NewAgentFromConfig(ctx, ac)
	if err != nil {
		return nil, err
	}

	s.agents = append(s.agents, *a)
	return a, nil
}
