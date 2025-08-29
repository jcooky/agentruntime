package engine

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/sliceutils"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func (p *ChatPromptValues) WithRecentConversations(conversations []Conversation) *ChatPromptValues {
	cloned := *p
	cloned.RecentConversations = conversations
	return &cloned
}

func (s *Engine) BuildPromptValues(ctx context.Context, agent entity.Agent, req RunRequest, summary *string) (*ChatPromptValues, error) {
	// construct inst promptValues
	promptValues := &ChatPromptValues{
		Agent:               agent,
		MessageExamples:     sliceutils.RandomSampleN(agent.MessageExamples, 100),
		RecentConversations: req.History,
		AvailableActions:    make([]AvailableAction, 0, len(agent.Skills)),
		Thread: Thread{
			Instruction:  req.ThreadInstruction,
			Participants: req.Participant,
			Files:        req.Files,
		},
		UserInfo: req.UserInfo,
		System:   agent.System,
	}

	// If we have a conversation summary, we need to extend the prompt values
	if summary != nil {
		// We'll handle this in template, but store the summary for now
		// This requires extending ChatPromptValues or handling it differently
		// For now, we'll add it to the system prompt
		promptValues.System += fmt.Sprintf("\n\n<conversation_summary>\n# Previous Conversation Summary\n%s\n</conversation_summary>", *summary)
	}

	// build available actions
	promptValues.Tools = make([]ai.Tool, 0, len(agent.Skills))
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

func convertToMessages(promptValues *ChatPromptValues) ([]*ai.Message, error) {
	var buf strings.Builder
	if err := chatInstTmpl.Execute(&buf, promptValues); err != nil {
		return nil, err
	}
	prompt := buf.String()

	return []*ai.Message{
		{
			Role: ai.RoleUser,
			Content: slices.Concat(
				[]*ai.Part{
					ai.NewTextPart(prompt),
				},
				lo.Map(promptValues.Thread.Files, func(f File, _ int) *ai.Part {
					return ai.NewMediaPart(f.ContentType, f.Data)
				}),
				[]*ai.Part{
					ai.NewTextPart(
						fmt.Sprintf(
							"<documents>Attached files:\n%s\n</documents>",
							strings.Join(lo.Map(promptValues.Thread.Files, func(f File, i int) string {
								return fmt.Sprintf("%d. filename:'%s', content_type:'%s', data_length:%d", i+1, f.Filename, f.ContentType, len(f.Data))
							}), "\n"),
						),
					),
				},
			),
		},
	}, nil
}
