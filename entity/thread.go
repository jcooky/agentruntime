package entity

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Thread struct {
	gorm.Model

	Instruction      string
	ParticipantNames datatypes.JSONSlice[string]

	Messages []Message `gorm:"foreignKey:ThreadID"`
}
