package network

import (
	"context"

	"github.com/habiliai/agentruntime/config"
	"github.com/jcooky/go-din"
	"github.com/ybbus/jsonrpc/v3"
)

type (
	JsonRpcClient interface {
		CheckLive(ctx context.Context, request *CheckLiveRequest) error
		GetAgentRuntimeInfo(ctx context.Context, request *GetAgentRuntimeInfoRequest) (*GetAgentRuntimeInfoResponse, error)
		RegisterAgent(ctx context.Context, request *RegisterAgentRequest) error
		DeregisterAgent(ctx context.Context, request *DeregisterAgentRequest) error
	}

	jsonRpcClient struct {
		client jsonrpc.RPCClient
	}
)

func NewJsonRpcClient(url string) JsonRpcClient {
	client := jsonrpc.NewClient(url)
	return &jsonRpcClient{
		client: client,
	}
}

func (c *jsonRpcClient) CheckLive(ctx context.Context, request *CheckLiveRequest) error {
	var reply struct{}
	err := c.client.CallFor(ctx, &reply, servicePrefix+".CheckLive", request)
	return err
}

func (c *jsonRpcClient) GetAgentRuntimeInfo(ctx context.Context, request *GetAgentRuntimeInfoRequest) (*GetAgentRuntimeInfoResponse, error) {
	var response GetAgentRuntimeInfoResponse
	err := c.client.CallFor(ctx, &response, servicePrefix+".GetAgentRuntimeInfo", request)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *jsonRpcClient) RegisterAgent(ctx context.Context, request *RegisterAgentRequest) error {
	var reply struct{}
	err := c.client.CallFor(ctx, &reply, servicePrefix+".RegisterAgent", request)
	return err
}

func (c *jsonRpcClient) DeregisterAgent(ctx context.Context, request *DeregisterAgentRequest) error {
	var reply struct{}
	err := c.client.CallFor(ctx, &reply, servicePrefix+".DeregisterAgent", request)
	return err
}

func init() {
	din.RegisterT(func(c *din.Container) (JsonRpcClient, error) {
		runtimeConfig := din.MustGetT[*config.RuntimeConfig](c)

		return NewJsonRpcClient(runtimeConfig.NetworkBaseUrl), nil
	})
}
