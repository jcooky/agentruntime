package entity

import (
	"github.com/habiliai/agentruntime/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Knowledge struct {
	gorm.Model

	AgentName string                             `gorm:"index:idx_knowledge_agent"`
	Content   string                             `gorm:"type:text"`
	Metadata  datatypes.JSONType[map[string]any] `gorm:"type:jsonb"`
	Embedding datatypes.JSONType[[]float32]      `gorm:"type:jsonb"`
}

func (k *Knowledge) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(k).Error, "failed to save knowledge")
}

func (k *Knowledge) Delete(db *gorm.DB) error {
	return errors.Wrapf(db.Delete(k).Error, "failed to delete knowledge")
}
