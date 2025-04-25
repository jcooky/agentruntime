package runtime

import (
	"context"

	"github.com/habiliai/agentruntime/internal/di"
)

type agentRuntimeServer struct {
	UnsafeAgentRuntimeServer

	runtime Service
}

func (a *agentRuntimeServer) Run(ctx context.Context, req *RunRequest) (*RunResponse, error) {
	agents, err := a.runtime.findAgentsByNames(req.AgentNames)
	if err != nil {
		return nil, err
	}
	if err = a.runtime.Run(ctx, uint(req.ThreadId), agents); err != nil {
		return nil, err
	}
	return &RunResponse{}, nil
}

var (
	_         AgentRuntimeServer = (*agentRuntimeServer)(nil)
	ServerKey                    = di.NewKey()
)

func init() {
	di.Register(ServerKey, func(c context.Context, container *di.Container) (any, error) {
		return &agentRuntimeServer{
			runtime: di.MustGet[Service](c, container, ServiceKey),
		}, nil
	})
}
