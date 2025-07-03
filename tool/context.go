package tool

import (
	"context"

	"github.com/habiliai/agentruntime/entity"
)

type Context struct {
	context.Context

	skill *entity.AgentSkill
}

func (c *Context) GetSkill() *entity.AgentSkill {
	return c.skill
}
