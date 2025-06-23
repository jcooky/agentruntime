# Testing Remote MCP Support

This document explains how to test Remote MCP support in agentruntime.

## Test Methods

### 1. Testing with Local MCP Servers

The simplest way is to run official MCP servers locally:

```bash
# 1. Run test agent
go run cmd/agentruntime/main.go examples/mcp/local-mcp-agent.yaml -p 3001

# 2. Run playground in another terminal
cd playground
yarn dev

# 3. Test at http://localhost:3000
# - Create new thread
# - Select test-mcp-agent
# - Send commands like "List files in /tmp directory"
```

### 2. Testing with Remote SSE Server

You can test remote MCP functionality using a Mock SSE server:

```bash
# 1. Run Mock SSE server (separate terminal)
node test/mock-sse-server.js

# 2. Run Remote SSE agent (separate terminal)
go run cmd/agentruntime/main.go examples/mcp/remote-sse-agent.yaml -p 3001

# 3. Test with Playground (separate terminal)
cd playground
yarn dev
```

Available test commands:

- "Echo this message: Hello World" - tests echo tool
- "What time is it?" - tests get_time tool

### 3. Real MCP Servers

Test with official MCP servers:

#### Filesystem Server

```bash
# Run directly without installation (using npx)
# Already configured in examples/mcp/local-mcp-agent.yaml
```

#### Brave Search Server (Requires Brave API key)

```bash
# 1. Get Brave API key: https://brave.com/search/api/
# 2. Uncomment brave-search section in local-mcp-agent.yaml
# 3. Replace BRAVE_API_KEY with actual key
```

#### GitHub Server (Requires GitHub token)

```bash
# 1. Create GitHub Personal Access Token
# 2. Uncomment github section in local-mcp-agent.yaml
# 3. Replace GITHUB_PERSONAL_ACCESS_TOKEN with actual token
```

## Test Scenarios

### Scenario 1: File System Operations

```
User: "Create a file called test.txt in /tmp with content 'Hello MCP'"
User: "List all files in /tmp directory"
User: "Read the content of /tmp/test.txt"
```

### Scenario 2: Echo Server (Remote SSE)

```
User: "Please echo this message: Testing Remote MCP"
User: "What's the current server time?"
```

### Scenario 3: Complex Operations

```
User: "Create a file with the current timestamp as content"
User: "Search for information about MCP protocol" (Requires Brave Search)
```

## Debugging

### 1. Check Server Logs

```bash
# Set log level when running agentruntime
LOG_LEVEL=debug go run cmd/agentruntime/main.go examples/mcp/remote-sse-agent.yaml
```

### 2. Mock Server Logs

The Mock SSE server outputs all requests and responses to the console.

### 3. Common Issues

**Connection Failed**

- Verify URL is correct
- Check if server is running
- Check firewall/port issues

**Authentication Failed**

- Verify API keys or tokens are correct
- Check Headers format

**Tool Not Found**

- Verify MCP server initialized properly
- Check tools/list response

## Advanced Testing

### OAuth Testing

OAuth testing requires a real OAuth provider. For testing purposes, you can implement a Mock OAuth server or use real services (e.g., Google, GitHub).

### Performance Testing

```bash
# Test with multiple MCP servers connected simultaneously
# Create examples/mcp/multi-mcp-agent.yaml with multiple skills
```

### Error Handling Testing

- Attempt connection with invalid URL
- Simulate network interruption
- Provide invalid authentication credentials
