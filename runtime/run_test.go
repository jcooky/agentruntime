package runtime_test

import (
	"time"

	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/network"
	"github.com/mokiat/gog"
	"github.com/stretchr/testify/mock"
)

func (s *AgentRuntimeTestSuite) TestRun() {
	var agents []entity.Agent
	for _, agentConfig := range s.agents {
		ag, err := s.runtime.RegisterAgent(s, agentConfig)
		s.Require().NoError(err)

		agents = append(agents, *ag)
	}

	threadId := uint32(1)
	s.agentNetwork.On("GetThread", mock.Anything, mock.MatchedBy(func(in *network.GetThreadRequest) bool {
		return in.ThreadId == threadId
	})).Return(&network.Thread{
		Id:          uint32(threadId),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Instruction: "# Mission: AI agents dialogue with user",
	}, nil).Once()
	s.agentNetwork.On("GetMessages", mock.Anything, mock.MatchedBy(func(in *network.GetMessagesRequest) bool {
		return in.ThreadId == threadId
	})).Return(&network.GetMessagesResponse{
		Messages: []*network.Message{
			{
				Id:      1,
				Content: "Hello, what is the weather today in Seoul?",
				Sender:  "USER",
			},
		},
		NextCursor: 1,
	}, nil).Once()
	s.agentNetwork.On("AddMessage", mock.Anything, mock.MatchedBy(func(in *network.AddMessageRequest) bool {
		s.T().Logf(">> AddMessage: %v\n", in)
		if !s.Len(in.ToolCalls, 2) {
			return false
		}
		toolCallNames := gog.Map(in.ToolCalls, func(tc *network.MessageToolCall) string {
			return tc.Name
		})
		return s.Contains(toolCallNames, "done_agent") &&
			s.Contains(toolCallNames, "get_weather") &&
			in.ThreadId == threadId
	})).Return(&network.AddMessageResponse{
		MessageId: uint32(1),
	}, nil).Once()
	s.agentNetwork.On("GetAgentRuntimeInfo", mock.Anything, mock.MatchedBy(func(in *network.GetAgentRuntimeInfoRequest) bool {
		return in.All
	})).Return(&network.GetAgentRuntimeInfoResponse{
		AgentRuntimeInfo: []*network.AgentRuntimeInfo{
			{
				Info: &network.AgentInfo{
					Name: "agent1",
				},
			},
		},
	}, nil).Once()
	defer s.agentNetwork.AssertExpectations(s.T())

	s.Require().NoError(s.runtime.Run(s, uint(threadId), agents))
}
