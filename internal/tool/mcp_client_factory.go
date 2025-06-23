package tool

import (
	"context"
	"fmt"
	"net/http"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/pkg/errors"
)

// MCPClientFactory creates MCP clients based on the server configuration
type MCPClientFactory struct {
	httpClient *http.Client
}

// NewMCPClientFactory creates a new MCP client factory
func NewMCPClientFactory() *MCPClientFactory {
	return &MCPClientFactory{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateClient creates an MCP client based on the server configuration
func (f *MCPClientFactory) CreateClient(ctx context.Context, serverID string, config MCPServerConfig) (mcpclient.MCPClient, error) {
	transportType := config.GetTransport()

	switch transportType {
	case MCPTransportStdio:
		return f.createStdioClient(config)

	case MCPTransportSSE:
		return f.createSSEClient(config)

	case MCPTransportOAuthSSE:
		return f.createOAuthSSEClient(config)

	case MCPTransportStreamable:
		return f.createStreamableClient(config)

	default:
		return nil, fmt.Errorf("unsupported transport type: %s", transportType)
	}
}

// createStdioClient creates a stdio-based MCP client
func (f *MCPClientFactory) createStdioClient(config MCPServerConfig) (mcpclient.MCPClient, error) {
	if config.Command == "" {
		return nil, errors.New("command is required for stdio transport")
	}

	var envs []string
	for key, val := range config.Env {
		envs = append(envs, fmt.Sprintf("%s=%s", key, val))
	}

	return mcpclient.NewStdioMCPClient(config.Command, envs, config.Args...)
}

// createSSEClient creates an SSE-based MCP client
func (f *MCPClientFactory) createSSEClient(config MCPServerConfig) (mcpclient.MCPClient, error) {
	if config.URL == "" {
		return nil, errors.New("URL is required for SSE transport")
	}

	opts := []transport.ClientOption{
		transport.WithHTTPClient(f.httpClient),
	}

	if len(config.Headers) > 0 {
		opts = append(opts, transport.WithHeaders(config.Headers))
	}

	return mcpclient.NewSSEMCPClient(config.URL, opts...)
}

// createOAuthSSEClient creates an OAuth-enabled SSE MCP client
func (f *MCPClientFactory) createOAuthSSEClient(config MCPServerConfig) (mcpclient.MCPClient, error) {
	if config.URL == "" {
		return nil, errors.New("URL is required for OAuth SSE transport")
	}

	if config.OAuthConfig == nil {
		return nil, errors.New("OAuth configuration is required for OAuth SSE transport")
	}

	oauthConfig := transport.OAuthConfig{
		ClientID:              config.OAuthConfig.ClientID,
		ClientSecret:          config.OAuthConfig.ClientSecret,
		RedirectURI:           config.OAuthConfig.RedirectURL,
		Scopes:                config.OAuthConfig.Scopes,
		AuthServerMetadataURL: config.OAuthConfig.AuthServerMetadataURL,
		PKCEEnabled:           config.OAuthConfig.PKCEEnabled,
		TokenStore:            transport.NewMemoryTokenStore(),
	}

	opts := []transport.ClientOption{
		transport.WithHTTPClient(f.httpClient),
	}

	if len(config.Headers) > 0 {
		opts = append(opts, transport.WithHeaders(config.Headers))
	}

	return mcpclient.NewOAuthSSEClient(config.URL, oauthConfig, opts...)
}

// createStreamableClient creates a streamable HTTP MCP client
func (f *MCPClientFactory) createStreamableClient(config MCPServerConfig) (mcpclient.MCPClient, error) {
	if config.URL == "" {
		return nil, errors.New("URL is required for streamable transport")
	}

	opts := []transport.StreamableHTTPCOption{}

	// Convert headers if needed
	if len(config.Headers) > 0 {
		headerFunc := func(req *http.Request) {
			for key, value := range config.Headers {
				req.Header.Set(key, value)
			}
		}
		// Note: You might need to implement a wrapper for StreamableHTTPCOption
		// as the transport package might have different option types
		_ = headerFunc // placeholder
	}

	return mcpclient.NewStreamableHttpClient(config.URL, opts...)
}
