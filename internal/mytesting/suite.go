package mytesting

import (
	"context"
	"github.com/jcooky/go-din"

	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	context.Context

	Container *din.Container
	Cancel    context.CancelFunc
}

func (s *Suite) SetupTest() {
	s.Context, s.Cancel = context.WithCancel(context.TODO())
	s.Container = din.NewContainer(s.Context, din.EnvTest)
}

func (s *Suite) TearDownTest() {
	s.Cancel()
}
