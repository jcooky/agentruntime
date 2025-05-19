package jsonrpc_test

import (
	"net/http/httptest"

	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/stretchr/testify/mock"
)

func (s *Suite) TestCreateThread() {
	// Given
	server := httptest.NewServer(s.handler)
	defer server.Close()

	s.network.On("CreateThread", mock.Anything, mock.MatchedBy(func(req *network.CreateThreadRequest) bool {
		return req.Instruction == "hello world" && req.Metadata["key"] == "value"
	})).Return(
		&network.CreateThreadResponse{
			ThreadId: 1,
		}, nil,
	).Once()

	client := network.NewJsonRpcClientWithHttpClient(server.URL+"/rpc", server.Client())
	// When
	resp, err := client.CreateThread(s, &network.CreateThreadRequest{
		Instruction: "hello world",
		Metadata: map[string]string{
			"key": "value",
		},
	})

	// Then
	s.Require().NoError(err)
	s.Equal(uint32(1), resp.ThreadId)
}

func (s *Suite) TestRegisterAgent() {
	// Given
	addr := "http://localhost:8080"
	agentInfo := []*network.AgentInfo{
		{
			Name: "agent1",
			Role: "role1",
		},
	}

	server := httptest.NewServer(s.handler)
	defer server.Close()

	s.network.On("RegisterAgent", mock.Anything, addr, agentInfo).Return(nil).Once()
	defer s.network.AssertExpectations(s.T())

	// When
	client := network.NewJsonRpcClientWithHttpClient(server.URL+"/rpc", server.Client())

	err := client.RegisterAgent(s, &network.RegisterAgentRequest{
		Addr: "http://localhost:8080",
		Info: []*network.AgentInfo{
			{
				Name: "agent1",
				Role: "role1",
			},
		},
	})

	// Then
	s.Require().NoError(err)
}

func (s *Suite) TestRunAgentRuntime() {
	// Given
	server := httptest.NewServer(s.handler)
	defer server.Close()

	s.runtime.On(
		"Run",
		mock.Anything,
		uint(1),
		mock.MatchedBy(func(agents []entity.Agent) bool {
			return len(agents) == 2 && agents[0].Name == "agent1" && agents[1].Name == "agent2"
		}),
	).Return(nil).Once()
	s.runtime.On("FindAgentsByNames", mock.Anything, mock.Anything).Return(
		[]entity.Agent{
			{

				Name: "agent1",
			},
			{
				Name: "agent2",
			},
		},
		nil,
	).Once()
	defer s.runtime.AssertExpectations(s.T())

	// When
	client := runtime.NewJsonRpcClientWithHttpClient(server.URL+"/rpc", server.Client())

	_, err := client.Run(s, &runtime.RunRequest{
		ThreadId:   1,
		AgentNames: []string{"agent1", "agent2"},
	})

	// Then
	s.Require().NoError(err)
}
