package mytesting

import (
	"context"

	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	context.Context

	Cancel context.CancelFunc
}

func (s *Suite) SetupTest() {
	s.Context, s.Cancel = context.WithCancel(context.TODO())
}

func (s *Suite) TearDownTest() {
	s.Cancel()
}
