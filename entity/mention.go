package entity

import (
	"time"

	"github.com/habiliai/agentruntime/errors"
	"gorm.io/gorm"
)

type Mention struct {
	AgentName string `gorm:"primarykey"`
	ThreadID  uint   `gorm:"primarykey"`
	Thread    Thread `gorm:"foreignKey:ThreadID"`

	CreatedAt time.Time
}

func (m *Mention) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(m).Error, "failed to save mention")
}

func (m *Mention) Delete(db *gorm.DB) error {
	return errors.Wrapf(db.Where("thread_id = ? AND agent_name = ?", m.ThreadID, m.AgentName).Delete(m).Error, "failed to delete mention")
}
