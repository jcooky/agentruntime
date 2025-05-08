package runtime

import (
	"context"
	"github.com/jcooky/go-din"
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
	_ AgentRuntimeServer = (*agentRuntimeServer)(nil)
)

func init() {
	din.RegisterT(func(c *din.Container) (AgentRuntimeServer, error) {
		return &agentRuntimeServer{
			runtime: din.MustGetT[Service](c),
		}, nil
	})
}
