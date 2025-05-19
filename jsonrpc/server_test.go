package jsonrpc_test

import (
	"net/http/httptest"

	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/stretchr/testify/mock"
)

func (s *Suite) TestCreateThread() {
	server := httptest.NewServer(s.handler)
	defer server.Close()

	client := network.NewJsonRpcClientWithHttpClient(server.URL+"/rpc", server.Client())

	resp, err := client.CreateThread(s, &network.CreateThreadRequest{
		Instruction: "hello world",
		Metadata: map[string]string{
			"key": "value",
		},
	})
	s.Require().NoError(err)
	s.Equal(uint32(1), resp.ThreadId)
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
