package engine

import (
	"context"
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
	}

	// build available actions
	promptValues.Tools = make([]ai.ToolRef, 0, len(agent.Skills))
	for _, skill := range agent.Skills {
		switch skill.Type {
		case "llm", "nativeTool":
			tool := s.toolManager.GetTool(skill.Name)
			if tool == nil {
				return nil, errors.Errorf("invalid tool name %s", skill.Name)
			}
			promptValues.AvailableActions = append(promptValues.AvailableActions, AvailableAction{
				Action:      skill.Name,
				Description: tool.Definition().Description,
			})
			promptValues.Tools = append(promptValues.Tools, tool)
		case "mcp":
			skillToolNames := skill.Tools
			if len(skillToolNames) == 0 {
				for _, tool := range s.toolManager.GetMCPTools(ctx, skill.Name) {
					skillToolNames = append(skillToolNames, tool.Name())
				}
			}
			for _, skillToolName := range skillToolNames {
				tool := s.toolManager.GetMCPTool(skill.Name, skillToolName)
				if tool == nil {
					return nil, errors.Errorf("invalid tool name %s", skill.Name)
				}
				promptValues.AvailableActions = append(promptValues.AvailableActions, AvailableAction{
					Action:      skillToolName,
					Description: tool.Definition().Description,
				})
				promptValues.Tools = append(promptValues.Tools, tool)
			}
		}
	}

	// Retrieve relevant knowledge using RAG
	promptValues.Knowledge = make([]string, 0)
	if len(history) > 0 {
		// Use the most recent conversation as context for knowledge retrieval
		lastConversation := history[len(history)-1]
		queryContext := lastConversation.Text
		if queryContext == "" && len(lastConversation.Actions) > 0 {
			// If no text, use action names as context
			actionNames := make([]string, len(lastConversation.Actions))
			for i, action := range lastConversation.Actions {
				actionNames[i] = action.Name
			}
			queryContext = strings.Join(actionNames, " ")
		}

		if queryContext != "" {
			retrievedKnowledge, err := s.memoryService.RetrieveRelevantKnowledge(ctx, agent.Name, queryContext, 5)
			if err != nil {
				s.logger.Warn("failed to retrieve relevant knowledge", "agent", agent.Name, "error", err)
			} else {
				promptValues.Knowledge = retrievedKnowledge
			}
		}
	}

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
