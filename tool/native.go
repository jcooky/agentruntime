package tool

import (
	"github.com/habiliai/agentruntime/entity"
	"github.com/pkg/errors"
)

func registerNativeTool[In any, Out any](m *manager, toolName, toolDescription string, skill *entity.NativeAgentSkill, fn func(ctx *Context, input In) (Out, error)) error {
	toolNames := m.nativeSkillToolNames[skill.Name]
	if len(toolNames) > 0 {
		return errors.Errorf("tool %s already registered", toolName)
	}

	registerLocalTool(m, toolName, toolDescription, skill, fn)
	m.nativeSkillToolNames[skill.Name] = append(m.nativeSkillToolNames[skill.Name], toolName)

	return nil
}

func (m *manager) registerNativeSkill(skill *entity.NativeAgentSkill) error {
	if skill.Name == "" {
		return errors.New("native tool name is required")
	}
	switch skill.Name {
	case "get_weather":
		m.registerGetWeatherTool(skill)
	case "web_search":
		m.registerWebSearchTool()
	case "knowledge_search":
		m.registerKnowledgeSearchTool(skill)
	case "rss":
		m.registerRSSSkill(skill)
	}

	return nil
}
