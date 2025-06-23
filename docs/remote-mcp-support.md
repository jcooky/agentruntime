# Remote MCP Support Documentation

## Overview

The agentruntime now supports remote MCP (Model Context Protocol) servers via multiple transport types, extending beyond the original stdio-based local process communication. This enables connecting to MCP servers hosted remotely over HTTP/SSE connections.

## Supported Transport Types

1. **stdio** (default for local processes)

   - Runs MCP server as a local subprocess
   - Communication via stdin/stdout
   - Original implementation, fully backward compatible

2. **sse** (Server-Sent Events)

   - HTTP-based connection to remote MCP servers
   - Supports custom headers for authentication
   - Auto-detected when URL is provided

3. **oauth-sse**

   - SSE transport with OAuth 2.0 authentication
   - Supports PKCE for public clients
   - Token management handled automatically

4. **streamable**
   - HTTP streaming transport
   - For servers that support streamable HTTP connections

## Configuration

### Using MCPServerConfig

The new `MCPServerConfig` struct supports all transport types:

```go
type MCPServerConfig struct {
    // For stdio transport
    Command string              `json:"command,omitempty"`
    Args    []string           `json:"args,omitempty"`
    Env     map[string]string  `json:"env,omitempty"`

    // For remote transports
    URL       string            `json:"url,omitempty"`
    Transport MCPTransportType  `json:"transport,omitempty"`
    Headers   map[string]string `json:"headers,omitempty"`

    // For OAuth
    OAuthConfig *OAuthConfig   `json:"oauthConfig,omitempty"`
}
```

### Auto-Detection

Transport type is automatically detected based on configuration:

- If `Command` is set → stdio
- If `URL` is set → sse (unless explicitly specified)

## Usage Examples

### Testing with Playground

The recommended way to test Remote MCP support is through the agentruntime playground:

1. **Configure your agent with MCP tools**:

   Create an agent configuration file with remote MCP skills. See `examples/mcp/mcp-remote-agent.yaml` for a complete example:

   ```yaml
   name: 'mcp-test-agent'
   description: 'Agent with remote MCP tools'
   model: 'claude-3-5-sonnet-20241022'
   instructions: 'You are a helpful assistant with access to remote MCP tools.'

   skills:
     # Remote SSE server with API key
     - type: mcp
       name: weather-service
       url: https://mcp.example.com/weather
       headers:
         Authorization: Bearer ${WEATHER_API_KEY}

     # OAuth-protected server
     - type: mcp
       name: data-service
       url: https://api.example.com/mcp
       transport: oauth-sse
       oauth:
         clientId: ${OAUTH_CLIENT_ID}
         clientSecret: ${OAUTH_CLIENT_SECRET}
         authServerMetadataUrl: https://auth.example.com/.well-known/openid-configuration
         scopes: ['mcp:read', 'mcp:write']
         pkceEnabled: true

     # Local MCP server (backward compatible)
     - type: mcp
       name: local-tools
       command: /usr/local/bin/mcp-server
       args: ['--config', 'config.json']
   ```

2. **Start the agentruntime server**:

   ```bash
   # From the project root
   go run cmd/agentruntime/main.go config/agents/ -p 3001
   ```

3. **Run the playground**:

   ```bash
   cd playground
   yarn install
   yarn dev
   ```

4. **Test your MCP tools**:
   - Open http://localhost:3000 in your browser
   - Create a new thread
   - Select your MCP-enabled agent
   - Send messages that would trigger the MCP tools
   - The agent will use the remote MCP tools to respond

### Agent Configuration

MCP tools are configured in the agent YAML file under the `skills` section. Each MCP skill can be either local (stdio) or remote (SSE/OAuth/Streamable).

#### Skill Fields

| Field       | Type     | Description                               | Required           |
| ----------- | -------- | ----------------------------------------- | ------------------ |
| `type`      | string   | Must be "mcp" for MCP tools               | Yes                |
| `name`      | string   | Unique identifier for the MCP server      | Yes                |
| `command`   | string   | Path to executable (stdio only)           | For stdio          |
| `args`      | []string | Command arguments (stdio only)            | No                 |
| `url`       | string   | Remote server URL                         | For remote         |
| `transport` | string   | "stdio", "sse", "oauth-sse", "streamable" | No (auto-detected) |
| `headers`   | map      | HTTP headers for authentication           | No                 |
| `oauth`     | object   | OAuth configuration                       | For oauth-sse      |
| `env`       | map      | Environment variables                     | No                 |

#### OAuth Configuration Fields

| Field                   | Type     | Description             |
| ----------------------- | -------- | ----------------------- |
| `clientId`              | string   | OAuth client ID         |
| `clientSecret`          | string   | OAuth client secret     |
| `authServerMetadataUrl` | string   | OAuth metadata endpoint |
| `redirectUrl`           | string   | OAuth redirect URL      |
| `scopes`                | []string | OAuth scopes            |
| `pkceEnabled`           | bool     | Enable PKCE flow        |

## Configuration File Format

See `config/mcp_examples.yaml` for complete examples of different transport configurations.

## Security Considerations

1. **API Keys**: Use environment variables for sensitive data like API keys
2. **OAuth**: Enable PKCE for public clients
3. **Headers**: Don't hardcode authentication tokens
4. **HTTPS**: Always use HTTPS for remote connections

**Note**: Environment variable substitution (e.g., `${API_KEY}`) is not automatically performed in agent YAML files. You should either:

- Use the `env` field to pass environment variables to stdio MCP servers
- Store non-sensitive configuration directly in the YAML
- Implement a secure configuration management solution for production deployments
- Consider using a secrets management service

## Migration Guide

To migrate from stdio to SSE:

1. Update your `RegisterMCPToolRequest` to use `ServerConfig`
2. Replace `Command` with `URL` pointing to your remote server
3. Add authentication headers if required
4. Test the connection

## Troubleshooting

### Common Issues

1. **Connection Failed**: Check URL and network connectivity
2. **Authentication Error**: Verify API keys/tokens
3. **Protocol Mismatch**: Ensure server supports MCP protocol version
4. **Timeout**: Adjust HTTP client timeout if needed

### Debug Tips

- Enable debug logging to see HTTP requests
- Check server logs for authentication issues
- Use curl to test SSE endpoints directly
- Verify OAuth configuration with provider documentation
