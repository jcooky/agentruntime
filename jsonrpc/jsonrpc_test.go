package jsonrpc_test

import (
	"net/http"
	"testing"

	"github.com/habiliai/agentruntime/internal/mytesting"
	"github.com/habiliai/agentruntime/jsonrpc"
	"github.com/habiliai/agentruntime/runtime"
	runtimetest "github.com/habiliai/agentruntime/runtime/test"
	"github.com/jcooky/go-din"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	mytesting.Suite

	handler http.Handler
	runtime *runtimetest.RuntimeServiceMock
}

func (s *Suite) SetupTest() {
	s.Suite.SetupTest()

	s.runtime = &runtimetest.RuntimeServiceMock{}
	din.SetT[runtime.Service](s.Container, s.runtime)
	s.handler = jsonrpc.NewHandlerWithHealth(s.Container, jsonrpc.WithNetwork(), jsonrpc.WithRuntime())
}

func (s *Suite) TearDownTest() {
	s.handler = nil
	s.Suite.TearDownTest()
}

func TestJsonRpc(t *testing.T) {
	suite.Run(t, new(Suite))
}
