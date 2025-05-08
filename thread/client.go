package thread

import (
	"github.com/jcooky/go-din"

	"github.com/habiliai/agentruntime/network"
	"google.golang.org/grpc"
)

func init() {
	din.RegisterT(func(c *din.Container) (ThreadManagerClient, error) {
		clientConn, err := din.Get[*grpc.ClientConn](c, network.GrpcClientConnKey)
		if err != nil {
			return nil, err
		}

		return NewThreadManagerClient(clientConn), nil
	})
}
