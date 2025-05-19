package jsonrpc_test

import (
	"net/http"
	"testing"

	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/habiliai/agentruntime/jsonrpc"
	"github.com/habiliai/agentruntime/network"
	networktest "github.com/habiliai/agentruntime/network/test"
	"github.com/habiliai/agentruntime/runtime"
	runtimetest "github.com/habiliai/agentruntime/runtime/test"
	"github.com/jcooky/go-din"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	mytesting.Suite

	handler http.Handler
	runtime *runtimetest.ServiceMock
	network *networktest.ServiceMock
}

func (s *Suite) SetupTest() {
	s.Suite.SetupTest()

	s.runtime = &runtimetest.ServiceMock{}
	din.SetT[runtime.Service](s.Container, s.runtime)
	s.network = &networktest.ServiceMock{}
	din.SetT[network.Service](s.Container, s.network)
	s.handler = jsonrpc.NewHandlerWithHealth(s.Container, jsonrpc.WithNetwork(), jsonrpc.WithRuntime())
}

func (s *Suite) TearDownTest() {
	s.handler = nil
	s.Suite.TearDownTest()
}

func TestJsonRpc(t *testing.T) {
	suite.Run(t, new(Suite))
}
