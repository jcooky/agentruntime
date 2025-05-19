package entity

import (
	"time"

	"github.com/habiliai/agentruntime/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AgentRuntime struct {
	gorm.Model

	Name       string `gorm:"index:idx_agent_name_uniq,unique,where:deleted_at IS NULL"`
	Role       string
	Addr       string
	LastLiveAt time.Time
	Metadata   datatypes.JSONType[map[string]string]
}

func (a *AgentRuntime) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(a).Error, "failed to save agent runtime")
}

func (a *AgentRuntime) Delete(db *gorm.DB) error {
	return errors.Wrapf(db.Delete(a).Error, "failed to delete agent runtime")
}
