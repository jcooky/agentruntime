package knowledge_test

import "github.com/habiliai/agentruntime/internal/mytesting"

type KnowledgeTestSuite struct {
	mytesting.Suite
}

func (s *KnowledgeTestSuite) SetupTest() {
	s.Suite.SetupTest()
}

func (s *KnowledgeTestSuite) TearDownTest() {
	s.Suite.TearDownTest()
}
