package engine

import (
	"context"
	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/entity"
	myerrors "github.com/habiliai/agentruntime/errors"
	"github.com/pkg/errors"
	"strings"
)

func (s *engine) NewAgentFromConfig(
	ctx context.Context,
	ac config.AgentConfig,
) (*entity.Agent, error) {
	var a entity.Agent

	a.Name = ac.Name
	a.System = ac.System
	a.Bio = ac.Bio
	a.Role = ac.Role
	a.Lore = ac.Lore
	a.MessageExamples = make([][]entity.MessageExample, 0, len(ac.MessageExamples))
	a.ModelName = ac.Model
	if a.ModelName == "" {
		a.ModelName = "gpt-4o"
	}
	for _, ex := range ac.MessageExamples {
		var messages []entity.MessageExample
		for _, msg := range ex.Messages {
			messages = append(messages, entity.MessageExample{
				User:    msg.Name,
				Text:    msg.Text,
				Actions: msg.Actions,
			})
		}
		a.MessageExamples = append(a.MessageExamples, messages)
	}

	for _, agentTool := range ac.Tools {
		toolNames := strings.SplitN(agentTool, "/", 2)
		if len(toolNames) == 1 {
			v := s.toolManager.GetTool(ctx, toolNames[0])
			if v == nil {
				return nil, errors.Wrapf(myerrors.ErrInvalidConfig, "invalid tool name %s", agentTool)
			}
			a.Tools = append(a.Tools, entity.Tool{
				Name:        v.Definition().Name,
				Description: v.Definition().Description,
			})
		} else if len(toolNames) == 2 {
			var tools []ai.Tool
			if toolNames[1] == "*" {
				tools = append(tools, s.toolManager.GetMCPTools(ctx, toolNames[0])...)
			} else {
				tools = append(tools, s.toolManager.GetMCPTool(ctx, toolNames[0], toolNames[1]))
			}
			for _, v := range tools {
				if v == nil {
					return nil, errors.Wrapf(myerrors.ErrInvalidConfig, "invalid tool name %s", agentTool)
				}
				a.Tools = append(a.Tools, entity.Tool{
					Name:        v.Definition().Name,
					Description: v.Definition().Description,
				})
			}
		} else {
			return nil, errors.Wrapf(myerrors.ErrInvalidConfig, "invalid tool name %s", agentTool)
		}
	}

	a.Metadata = ac.Metadata
	a.Knowledge = ac.Knowledge

	return &a, nil
}
