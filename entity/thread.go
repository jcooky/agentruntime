package entity

import (
	"gorm.io/gorm"
)

type Thread struct {
	gorm.Model

	Instruction  string
	Participants []AgentRuntime `gorm:"many2many:thread_participants;"`

	Messages []Message `gorm:"foreignKey:ThreadID"`
}
