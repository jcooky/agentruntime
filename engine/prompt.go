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

func (s *Engine) BuildPromptValues(ctx context.Context, agent entity.Agent, req RunRequest) (*ChatPromptValues, error) {
	var recentConversations []Conversation
	var conversationSummary *SummarizedConversation

	// Use conversation summarizer if available
	if s.conversationSummarizer != nil && len(req.History) > 0 {
		result, err := s.conversationSummarizer.ProcessConversationHistory(ctx, req.History, req.Files)
		if err != nil {
			s.logger.Error("failed to process conversation history", "error", err)
			// Fall back to simple truncation
			recentConversations = sliceutils.Cut(req.History, -200, len(req.History))
		} else {
			recentConversations = result.RecentConversations
			conversationSummary = result.Summary
		}
	} else {
		// Fall back to simple truncation when summarizer is not available
		recentConversations = sliceutils.Cut(req.History, -200, len(req.History))
	}

	// construct inst promptValues
	promptValues := &ChatPromptValues{
		Agent:               agent,
		MessageExamples:     sliceutils.RandomSampleN(agent.MessageExamples, 100),
		RecentConversations: recentConversations,
		AvailableActions:    make([]AvailableAction, 0, len(agent.Skills)),
		Thread: Thread{
			Instruction:  req.ThreadInstruction,
			Participants: req.Participant,
		},
		UserInfo: req.UserInfo,
		System:   agent.System,
	}

	// If we have a conversation summary, we need to extend the prompt values
	if conversationSummary != nil {
		// We'll handle this in template, but store the summary for now
		// This requires extending ChatPromptValues or handling it differently
		// For now, we'll add it to the system prompt
		promptValues.System += fmt.Sprintf("\n\n<conversation_summary>\n# Previous Conversation Summary\n%s\n</conversation_summary>", conversationSummary.Summary)
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
