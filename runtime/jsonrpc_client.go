package runtime

import (
	"context"

	"github.com/ybbus/jsonrpc/v3"
)

type (
	JsonRpcClient interface {
		Run(ctx context.Context, req *RunRequest) (*RunResponse, error)
	}

	jsonRpcClient struct {
		client jsonrpc.RPCClient
	}
)

func (c *jsonRpcClient) Run(ctx context.Context, req *RunRequest) (*RunResponse, error) {
	var reply RunResponse
	err := c.client.CallFor(ctx, &reply, servicePrefix+".Run", req)
	if err != nil {
		return nil, err
	}
	return &reply, nil
}

func NewJsonRpcClient(url string) JsonRpcClient {
	client := jsonrpc.NewClient(url)
	return &jsonRpcClient{
		client: client,
	}
}
