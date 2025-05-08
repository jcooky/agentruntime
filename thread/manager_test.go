package thread_test

import (
	"github.com/habiliai/agentruntime/internal/db"
	"github.com/jcooky/go-din"
	"gorm.io/gorm"
	"os"
	"testing"

	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/habiliai/agentruntime/thread"
	"github.com/stretchr/testify/suite"
	"gorm.io/datatypes"
)

type ThreadManagerTestSuite struct {
	mytesting.Suite

	threadManager thread.Manager
	DB            *gorm.DB
}

func (s *ThreadManagerTestSuite) SetupTest() {
	os.Setenv("ENV_TEST_FILE", "../.env.test")
	s.Suite.SetupTest()

	s.threadManager = din.MustGetT[thread.Manager](s.Container)
	s.DB = din.MustGet[*gorm.DB](s.Container, db.Key)
}

func (s *ThreadManagerTestSuite) TearDownTest() {
	s.Suite.TearDownTest()
}

func (s *ThreadManagerTestSuite) TestGetMessagesInThread() {
	thread := entity.Thread{
		Instruction: "",
	}
	s.Require().NoError(s.DB.Save(&thread).Error)

	messages := []entity.Message{
		{ThreadID: thread.ID, Content: datatypes.NewJSONType(entity.MessageContent{Text: "Message 1"}), User: "USER"},
		{ThreadID: thread.ID, Content: datatypes.NewJSONType(entity.MessageContent{Text: "Message 2"}), User: "Sunny"},
		{ThreadID: thread.ID, Content: datatypes.NewJSONType(entity.MessageContent{Text: "Message 3"}), User: "Eric"},
		{ThreadID: thread.ID, Content: datatypes.NewJSONType(entity.MessageContent{Text: "Message 4"}), User: "USER"},
	}

	for _, msg := range messages {
		s.Require().NoError(s.DB.Save(&msg).Error)
	}

	resp, err := s.threadManager.GetMessages(s.Context, thread.ID, "ASC", 0, 100)
	s.Require().NoError(err)

	s.Require().Equal(len(messages), len(resp))
}

func TestThreadManager(t *testing.T) {
	suite.Run(t, new(ThreadManagerTestSuite))
}
