package tool

import (
	"context"

	"github.com/habiliai/agentruntime/entity"
)

type Context struct {
	context.Context

	skill *entity.NativeAgentSkill
}

func (c *Context) GetSkill() *entity.NativeAgentSkill {
	return c.skill
}
