package tool

import (
	"context"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/myctx"
	"github.com/pkg/errors"
)

type (
	DoneAgentRequest struct {
		Reason string `json:"reason" jsonschema:"description=Reason why the task is considered done"`
	}

	DoneAgentResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
)

func (m *manager) DoneAgent(ctx context.Context, req *DoneAgentRequest) (*DoneAgentResponse, error) {
	// Verify thread exists
	thread := myctx.GetThreadFromContext(ctx)
	if thread == nil {
		return nil, errors.New("thread not found in context")
	}

	agent := myctx.GetAgentFromContext(ctx)
	if agent == nil {
		return nil, errors.New("agent not found in context")
	}

	return &DoneAgentResponse{
		Success: true,
		Message: "Task marked as completed: " + req.Reason,
	}, nil
}

func init() {
	RegisterLocalTool(
		"done_agent",
		"Mark the current task as completed when you've fulfilled all requirements",
		func(ctx context.Context, req struct {
			*DoneAgentRequest
		}) (res struct {
			*DoneAgentResponse
		}, err error) {
			s := di.MustGet[*manager](ctx, ManagerKey)
			res.DoneAgentResponse, err = s.DoneAgent(ctx, req.DoneAgentRequest)
			return
		},
	)
}
