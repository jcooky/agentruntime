package network

import (
	"context"
	"net/http"

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
		GetMessages(ctx context.Context, request *GetMessagesRequest) (*GetMessagesResponse, error)
		GetNumMessages(ctx context.Context, request *GetNumMessagesRequest) (*GetNumMessagesResponse, error)
		CreateThread(ctx context.Context, request *CreateThreadRequest) (*CreateThreadResponse, error)
		GetThread(ctx context.Context, request *GetThreadRequest) (*Thread, error)
		AddMessage(ctx context.Context, request *AddMessageRequest) (*AddMessageResponse, error)
		IsMentionedOnce(ctx context.Context, request *IsMentionedRequest) (*IsMentionedResponse, error)
	}

	jsonRpcClient struct {
		client jsonrpc.RPCClient
	}
)

func NewJsonRpcClient(url string) JsonRpcClient {
	return NewJsonRpcClientWithHttpClient(url, http.DefaultClient)
}

func NewJsonRpcClientWithHttpClient(url string, httpClient *http.Client) JsonRpcClient {
	client := jsonrpc.NewClientWithOpts(url, &jsonrpc.RPCClientOpts{
		HTTPClient: httpClient,
	})
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
	_, err := c.client.Call(ctx, servicePrefix+".RegisterAgent", request)
	return err
}

func (c *jsonRpcClient) DeregisterAgent(ctx context.Context, request *DeregisterAgentRequest) error {
	_, err := c.client.Call(ctx, servicePrefix+".DeregisterAgent", request)
	return err
}

func (c *jsonRpcClient) GetMessages(ctx context.Context, request *GetMessagesRequest) (*GetMessagesResponse, error) {
	var response GetMessagesResponse
	err := c.client.CallFor(ctx, &response, servicePrefix+".GetMessages", request)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *jsonRpcClient) GetNumMessages(ctx context.Context, request *GetNumMessagesRequest) (*GetNumMessagesResponse, error) {
	var response GetNumMessagesResponse
	err := c.client.CallFor(ctx, &response, servicePrefix+".GetNumMessages", request)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *jsonRpcClient) CreateThread(ctx context.Context, request *CreateThreadRequest) (*CreateThreadResponse, error) {
	var response CreateThreadResponse
	err := c.client.CallFor(ctx, &response, servicePrefix+".CreateThread", request)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *jsonRpcClient) GetThread(ctx context.Context, request *GetThreadRequest) (*Thread, error) {
	var response Thread
	err := c.client.CallFor(ctx, &response, servicePrefix+".GetThread", request)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *jsonRpcClient) AddMessage(ctx context.Context, request *AddMessageRequest) (*AddMessageResponse, error) {
	var response AddMessageResponse
	err := c.client.CallFor(ctx, &response, servicePrefix+".AddMessage", request)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *jsonRpcClient) IsMentionedOnce(ctx context.Context, request *IsMentionedRequest) (*IsMentionedResponse, error) {
	var response IsMentionedResponse
	err := c.client.CallFor(ctx, &response, servicePrefix+".IsMentionedOnce", request)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func init() {
	din.RegisterT(func(c *din.Container) (JsonRpcClient, error) {
		runtimeConfig := din.MustGetT[*config.RuntimeConfig](c)

		return NewJsonRpcClient(runtimeConfig.NetworkBaseUrl), nil
	})
}
