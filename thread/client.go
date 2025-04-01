package thread

import (
	"context"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/network"
	"google.golang.org/grpc"
)

var (
	ClientKey = di.NewKey()
)

func init() {
	di.Register(ClientKey, func(ctx context.Context, _ di.Env) (any, error) {
		clientConn, err := di.Get[*grpc.ClientConn](ctx, network.GrpcClientConnKey)
		if err != nil {
			return nil, err
		}

		return NewThreadManagerClient(clientConn), nil
	})
}
