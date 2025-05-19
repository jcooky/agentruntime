package runtime

import (
	"context"
	"net/http"

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
