package network

import (
	"github.com/jcooky/go-din"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/grpcutils"
	"google.golang.org/grpc"
)

var (
	GrpcClientConnKey = din.NewRandomName()
)

func init() {
	din.Register(GrpcClientConnKey, func(c *din.Container) (any, error) {
		conf, err := din.GetT[*config.RuntimeConfig](c)
		if err != nil {
			return nil, err
		}

		client, err := grpcutils.NewClient(conf.NetworkGrpcAddr, conf.NetworkGrpcSecure)
		if err != nil {
			return nil, err
		}

		go func() {
			<-c.Done()
			client.Close()
		}()

		return client, nil
	})
	din.RegisterT(func(c *din.Container) (AgentNetworkClient, error) {
		clientConn, err := din.Get[*grpc.ClientConn](c, GrpcClientConnKey)
		if err != nil {
			return nil, err
		}

		return NewAgentNetworkClient(clientConn), nil
	})
}
