package runtime_test

import (
	"io"

	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/thread"
	threadtest "github.com/habiliai/agentruntime/thread/test"
	"github.com/mokiat/gog"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AgentRuntimeTestSuite) TestRun() {
	var agents []entity.Agent
	for _, agentConfig := range s.agents {
		ag, err := s.runtime.RegisterAgent(s, agentConfig)
		s.Require().NoError(err)

		agents = append(agents, *ag)
	}

	getMessagesStreamMock := &threadtest.MockGetMessagesClient{}
	getMessagesStreamMock.On("Recv").Return(&thread.GetMessagesResponse{
		Messages: []*thread.Message{
			{
				Id:      1,
				Content: "Hello, what is the weather today in Seoul?",
				Sender:  "USER",
			},
		},
	}, nil).Once()
	getMessagesStreamMock.On("Recv").Return(&thread.GetMessagesResponse{}, io.EOF).Once()
	defer getMessagesStreamMock.AssertExpectations(s.T())

	threadId := uint32(1)
	s.threadManager.On("GetThread", mock.Anything, mock.MatchedBy(func(in *thread.GetThreadRequest) bool {
		return in.ThreadId == threadId
	})).Return(&thread.Thread{
		Id:          uint32(threadId),
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
		Instruction: "# Mission: AI agents dialogue with user",
	}, nil).Once()
	s.threadManager.On("GetMessages", mock.Anything, mock.MatchedBy(func(in *thread.GetMessagesRequest) bool {
		return in.ThreadId == threadId
	})).Return(getMessagesStreamMock, nil).Once()
	s.threadManager.On("AddMessage", mock.Anything, mock.MatchedBy(func(in *thread.AddMessageRequest) bool {
		s.T().Logf(">> AddMessage: %v\n", in)
		if !s.Len(in.ToolCalls, 2) {
			return false
		}
		toolCallNames := gog.Map(in.ToolCalls, func(tc *thread.Message_ToolCall) string {
			return tc.Name
		})
		return s.Contains(toolCallNames, "done_agent") &&
			s.Contains(toolCallNames, "get_weather") &&
			in.ThreadId == threadId
	})).Return(&thread.AddMessageResponse{
		MessageId: uint32(1),
	}, nil).Once()
	defer s.threadManager.AssertExpectations(s.T())
	s.agentNetwork.On("GetAgentRuntimeInfo", mock.Anything, mock.MatchedBy(func(in *network.GetAgentRuntimeInfoRequest) bool {
		return in.All != nil && *in.All
	})).Return(&network.GetAgentRuntimeInfoResponse{
		AgentRuntimeInfo: []*network.AgentRuntimeInfo{
			{
				Info: &network.AgentInfo{
					Name: "agent1",
				},
			},
		},
	}, nil).Once()

	s.Require().NoError(s.runtime.Run(s, uint(threadId), agents))
}
