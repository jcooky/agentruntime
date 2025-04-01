package err

import "fmt"

var (
	ErrInvalidConfig = fmt.Errorf("agentruntime: invalid config")
	ErrNotFound      = fmt.Errorf("agentruntime: not found")
	ErrNoMore        = fmt.Errorf("agentruntime: no more")
	ErrInvalidParams = fmt.Errorf("agentruntime: invalid params")
	ErrTimeout       = fmt.Errorf("agentruntime: timeout")
)
