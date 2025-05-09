package errors

import (
	"github.com/pkg/errors"
)

var (
	Wrapf     = errors.Wrapf
	Errorf    = errors.Errorf
	New       = errors.New
	WithStack = errors.WithStack
	Is        = errors.Is
	As        = errors.As
)
