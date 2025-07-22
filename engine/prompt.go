package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/sliceutils"
	"github.com/pkg/errors"
)

func (s *Engine) BuildPromptValues(ctx context.Context, agent entity.Agent, history []Conversation, thread Thread) (*ChatPromptValues, error) {

	// construct inst promptValues
	promptValues := &ChatPromptValues{
		Agent:               agent,
		MessageExamples:     sliceutils.RandomSampleN(agent.MessageExamples, 100),
		RecentConversations: sliceutils.Cut(history, -200, len(history)),
		AvailableActions:    make([]AvailableAction, 0, len(agent.Skills)),
		Thread: Thread{
			Instruction:  thread.Instruction,
			Participants: thread.Participants,
		},
		System: agent.System,
	}

	// build available actions
	promptValues.Tools = make([]ai.ToolRef, 0, len(agent.Skills))
	for _, skill := range agent.Skills {
		tools, err := s.toolManager.GetToolsBySkill(ctx, skill)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get tools by skill")
		}
		for _, tool := range tools {
			promptValues.AvailableActions = append(promptValues.AvailableActions, AvailableAction{
				Action:      tool.Name(),
				Description: tool.Definition().Description,
			})
			promptValues.Tools = append(promptValues.Tools, tool)

		}

		usagePrompt := s.toolManager.GetUsagePrompt(skill)
		if usagePrompt != "" {
			promptValues.System += fmt.Sprintf("\n\n%s", usagePrompt)
		}
	}

	promptValues.System = strings.TrimSpace(promptValues.System)

	return promptValues, nil
}

func GetPromptFn(promptValues *ChatPromptValues) ai.PromptFn {
	return func(ctx context.Context, _ any) (string, error) {
		var buf strings.Builder
		if err := chatInstTmpl.Execute(&buf, promptValues); err != nil {
			return "", err
		}
		result := buf.String()
		return result, nil
	}
}
