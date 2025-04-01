package entity

import (
	"gorm.io/gorm"
)

type Thread struct {
	gorm.Model

	Instruction string

	Messages []Message `gorm:"foreignKey:ThreadID"`
}
