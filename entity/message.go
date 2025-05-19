package entity

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model

	ThreadID uint
	Thread   Thread `gorm:"foreignKey:ThreadID"`

	User    string
	Content datatypes.JSONType[MessageContent]
}

type MessageContent struct {
	Text      string                   `json:"text,omitempty"`
	ToolCalls []MessageContentToolCall `json:"tool_calls,omitempty"`
	Error     string                   `json:"error,omitempty"`
}

type MessageContentToolCall struct {
	Name      string `json:"name,omitempty"`
	Arguments any    `json:"arguments,omitempty"`
	Result    any    `json:"result,omitempty"`
}
