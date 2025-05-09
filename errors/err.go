package errors

import (
	"fmt"
)

var (
	ErrInvalidConfig  = fmt.Errorf("agentruntime: invalid config")
	ErrNotFound       = fmt.Errorf("agentruntime: not found")
	ErrNoMore         = fmt.Errorf("agentruntime: no more")
	ErrInvalidParams  = fmt.Errorf("agentruntime: invalid params")
	ErrInternal       = fmt.Errorf("agentruntime: internal error")
	ErrInvalidRequest = fmt.Errorf("agentruntime: invalid request")
)
