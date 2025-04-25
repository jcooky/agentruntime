package network

import (
	"context"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/grpcutils"
	"google.golang.org/grpc"
)

var (
	GrpcClientConnKey = di.NewKey()
	ClientKey         = di.NewKey()
)

func init() {
	di.Register(GrpcClientConnKey, func(ctx context.Context, c *di.Container) (any, error) {
		conf, err := di.Get[*config.RuntimeConfig](ctx, c, config.RuntimeConfigKey)
		if err != nil {
			return nil, err
		}

		client, err := grpcutils.NewClient(conf.NetworkGrpcAddr, conf.NetworkGrpcSecure)
		if err != nil {
			return nil, err
		}

		go func() {
			<-ctx.Done()
			client.Close()
		}()

		return client, nil
	})
	di.Register(ClientKey, func(ctx context.Context, c *di.Container) (any, error) {
		clientConn, err := di.Get[*grpc.ClientConn](ctx, c, GrpcClientConnKey)
		if err != nil {
			return nil, err
		}

		return NewAgentNetworkClient(clientConn), nil
	})
}
