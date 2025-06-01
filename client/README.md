# AgentNetwork Client

A TypeScript client library for connecting to AgentNetwork server, designed for use in web applications and Node.js environments.

## Overview

The AgentNetwork Client is a TypeScript library that enables seamless communication with AgentNetwork servers using the JSON-RPC 2.0 protocol. This client provides a simple and type-safe interface for managing agents, threads, and messages in distributed agent systems.

## Features

- **TypeScript Support**: Full type definitions for a better development experience
- **JSON-RPC 2.0 Protocol**: Standards-compliant communication with AgentNetwork servers
- **Cross-Platform**: Works in both browser environments and Node.js applications
- **Comprehensive API**: Complete coverage of AgentNetwork server functionality
- **Async/Await Support**: Modern promise-based API for all operations
- **Error Handling**: Robust error handling with detailed error messages

## Installation

```bash
# Using npm
npm install @habiliai/agentnetwork-client

# Using yarn
yarn add @habiliai/agentnetwork-client

# Using pnpm
pnpm add @habiliai/agentnetwork-client
```

## Quick Start

```typescript
import { AgentNetworkClient } from '@habiliai/agentnetwork-client';

// Initialize the client
const client = new AgentNetworkClient({
  rpcEndpoint: 'http://localhost:8080/rpc',
});

// Register an agent
await client.RegisterAgent({
  addr: 'http://my-agent:8080',
  info: [
    {
      name: 'my-agent',
      description: 'My Agent',
      instructions: 'You are a helpful assistant.',
    },
  ],
});

// Create a thread
const { thread_id } = await client.CreateThread({
  instruction: 'Help me with programming tasks',
  participants: ['my-agent'],
});

// Send a message
const { message_id } = await client.AddMessage({
  thread_id: thread_id,
  sender: 'user',
  content: 'Hello, can you help me?',
  tool_calls: [],
});
```

## API Reference

### Constructor

```typescript
const client = new AgentNetworkClient(options: AgentNetworkClientOptions);
```

**Options:**

- `rpcEndpoint`: The URL of the AgentNetwork server RPC endpoint

### Methods

#### RegisterAgent

Register a new agent with the network.

```typescript
await client.RegisterAgent({
  addr: string;   // Agent's network address
  info: Array<{
    name: string;
    description: string;
    instructions: string;
  }>;
});
```

#### CheckLive

Check if agents are currently active.

```typescript
await client.CheckLive({ names: string[] });
```

#### GetAgentRuntimeInfo

Get runtime information for registered agents.

```typescript
const { agent_runtime_info } = await client.GetAgentRuntimeInfo({
  names?: string[];  // Optional: specific agent names
  all?: boolean;     // Optional: get all agents
});
```

#### CreateThread

Create a new conversation thread.

```typescript
const { thread_id } = await client.CreateThread({
  instruction: string;
  participants: string[];
  metadata?: Record<string, string>;
});
```

#### GetThread

Retrieve thread information by ID.

```typescript
const thread = await client.GetThread({ thread_id: number });
```

#### AddMessage

Add a message to a thread.

```typescript
const { message_id } = await client.AddMessage({
  thread_id: number;
  sender: string;
  content: string;
  tool_calls: Array<{
    id: string;
    name: string;
    arguments: Record<string, any>;
  }>;
});
```

#### GetMessages

Retrieve messages from a thread.

```typescript
const { messages, next_cursor } = await client.GetMessages({
  thread_id: number;
  order: 'latest' | 'oldest';
  cursor?: number;
  limit?: number;
});
```

#### GetNumMessages

Get the total number of messages in a thread.

```typescript
const { num_messages } = await client.GetNumMessages({ thread_id: number });
```

#### IsMentionedOnce

Check which threads have mentioned a specific agent exactly once.

```typescript
const { thread_ids } = await client.IsMentionedOnce({
  agent_name: string;
});
```

#### DeregisterAgent

Remove agents from the network.

```typescript
await client.DeregisterAgent({ names: string[] });
```

#### GetThreads

Get a list of threads with pagination.

```typescript
const { threads, next_cursor } = await client.GetThreads({
  cursor?: number;
  limit?: number;
});
```

## Advanced Usage

### Error Handling

```typescript
try {
  const thread = await client.CreateThread({
    instruction: 'Process this task',
    participants: ['agent-1', 'agent-2'],
  });
} catch (error) {
  if (error.code === -32600) {
    console.error('Invalid request:', error.message);
  } else {
    console.error('Unexpected error:', error);
  }
}
```

### Tool Calls

```typescript
const { message_id } = await client.AddMessage({
  thread_id: thread.thread_id,
  sender: 'user',
  content: 'Translate this text',
  tool_calls: [
    {
      id: 'call_123',
      name: 'translate',
      arguments: {
        text: 'Hello world',
        targetLanguage: 'es',
      },
    },
  ],
});
```

### Pagination

```typescript
// Get messages with pagination
const { messages, next_cursor } = await client.GetMessages({
  thread_id: thread.thread_id,
  order: 'latest',
  limit: 50,
});

// Get next page
if (next_cursor) {
  const nextPage = await client.GetMessages({
    thread_id: thread.thread_id,
    order: 'latest',
    cursor: next_cursor,
    limit: 50,
  });
}
```

## Requirements

- Node.js 16.0.0 or higher (for Node.js applications)
- Modern browser with ES2017 support (for web applications)
- AgentNetwork server running and accessible

## Protocol Details

This client communicates with AgentNetwork servers using the JSON-RPC 2.0 protocol. All requests are sent as HTTP POST requests with `Content-Type: application/json`.

### Request Format

```json
{
  "jsonrpc": "2.0",
  "method": "habiliai-agentnetwork-v1.MethodName",
  "params": { ... },
  "id": "unique-request-id"
}
```

### Response Format

```json
{
  "jsonrpc": "2.0",
  "result": { ... },
  "id": "unique-request-id"
}
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](../CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

## Support

For issues, questions, or contributions, please visit our [GitHub repository](https://github.com/habiliai/agentnetwork).
