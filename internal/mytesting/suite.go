package mytesting

import (
	"context"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/db"
	di "github.com/habiliai/agentruntime/internal/di"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Suite struct {
	suite.Suite
	context.Context

	Config    *config.RuntimeConfig
	DB        *gorm.DB
	Container *di.Container
	Cancel    context.CancelFunc
	eg        errgroup.Group
}

func (s *Suite) SetupTest() {
	s.Context, s.Cancel = context.WithCancel(context.TODO())
	s.Container = di.NewContainer(di.EnvTest)

	s.Config = di.MustGet[*config.RuntimeConfig](s.Context, s.Container, config.RuntimeConfigKey)
	s.DB = di.MustGet[*gorm.DB](s.Context, s.Container, db.Key)
}

func (s *Suite) TearDownTest() {
	if err := db.CloseDB(s.DB); err != nil {
		s.T().Logf("failed to close db: %v", err)
	}
	s.Cancel()
	s.eg.Wait()
}
