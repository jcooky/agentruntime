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
	di.Register(GrpcClientConnKey, func(ctx context.Context, env di.Env) (any, error) {
		conf, err := di.Get[*config.RuntimeConfig](ctx, config.RuntimeConfigKey)
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
	di.Register(ClientKey, func(ctx context.Context, _ di.Env) (any, error) {
		clientConn, err := di.Get[*grpc.ClientConn](ctx, GrpcClientConnKey)
		if err != nil {
			return nil, err
		}

		return NewAgentNetworkClient(clientConn), nil
	})
}
