# MCP (Model Context Protocol) Examples

This directory contains example configurations for using MCP servers with agentruntime.

## Files

### Agent Examples

- **`local-mcp-agent.yaml`** - Example agent using local MCP servers (filesystem, Brave search, GitHub)
- **`remote-sse-agent.yaml`** - Example agent using remote SSE-based MCP server
- **`mcp-remote-agent.yaml`** - Complete example showing various remote MCP configurations (SSE, OAuth, Streamable)

### Configuration Reference

- **`configuration-examples.yaml`** - Reference examples for different MCP server configuration formats

## Usage

To use these examples:

1. Copy the desired example to your agent configuration directory
2. Replace placeholder values with your actual credentials
3. Run the agent:

```bash
go run cmd/agentruntime/main.go examples/mcp/local-mcp-agent.yaml
```

## Testing

For testing instructions, see [Testing Remote MCP Support](../../docs/testing-remote-mcp.md).

## Notes

- Environment variable substitution (${VAR}) is not automatically supported in YAML files
- For production use, implement a secure configuration management solution
- Always use HTTPS for remote MCP connections
