<p align="center">
  <img alt="Shows a white agents.json Logo with a black background." src="https://u6mo491ntx4iwuoz.public.blob.vercel-storage.com/logo/bg_black_logo-tzo7s5eNJEWkXMEVBMME7ucb7BUN2L.png" width="full">
</p>

<h1 align="center">AI Agent Runtime by <i>Habili.ai</i> </h1>

[![Go Build & Test Pipeline](https://github.com/habiliai/agentruntime/actions/workflows/ci.yml/badge.svg)](https://github.com/habiliai/agentruntime/actions/workflows/ci.yml)
[![Go Lint Pipeline](https://github.com/habiliai/agentruntime/actions/workflows/lint.yml/badge.svg)](https://github.com/habiliai/agentruntime/actions/workflows/lint.yml)

## Overview

`agentruntime` is a comprehensive platform for building and running AI agents locally. It provides a flexible runtime that enables developers to create LLM-powered agents with various capabilities and tools through a simple, unified API.

### Key Features

- **Direct Agent Execution**: Use the `AgentRuntime` type to directly execute agents in your Go applications
- **Simple Agent Configuration**: Define agent capabilities, tools, and behavior through intuitive YAML configuration
- **Genkit Integration**: Seamlessly integrate with the Genkit platform for enhanced AI capabilities
- **Built-in Playground**: Test and interact with your agents through a modern web interface
- **Tool Extensibility**: Easily extend agent capabilities with custom tools and integrations
- **Multiple LLM Support**: Works with OpenAI, Anthropic, xAI and other providers
- **MCP Server Support**: Integrate with Model Context Protocol (MCP) servers for extended functionality

### RAG (Retrieval-Augmented Generation) Support

AgentRuntime includes built-in RAG functionality using [sqlite-vec](https://github.com/asg017/sqlite-vec) for high-performance vector operations combined with GORM entities for structured data management.

#### Features

- **Automatic Knowledge Indexing**: Agent knowledge is automatically indexed into vector embeddings when agents are created
- **High-Performance Vector Search**: Fast similarity search using sqlite-vec native SQLite extension
- **Structured Data Storage**: GORM entities for knowledge management with JSONB metadata
- **Context-Aware Retrieval**: Relevant knowledge is retrieved based on conversation context using vector similarity
- **OpenAI Embeddings**: Uses OpenAI's text-embedding-3-small model via genkit for consistent embeddings

#### Architecture

The RAG system uses sqlite-vec for all vector operations:

1. **GORM Entities**: Knowledge is stored in the `Knowledge` entity with full metadata and embeddings
2. **sqlite-vec Virtual Table**: Vector embeddings are stored in a virtual table for fast similarity search
3. **Integrated Operations**: Both storage systems work together in transactional safety

#### Quick Setup

1. **Configure agent with knowledge**:

   ```yaml
   name: TravelAgent
   model: openai/gpt-4o
   knowledge:
     - cityName: Seoul
       aliases: Seoul, SEOUL, KOR, Korea
       info: Capital city of South Korea, known for technology and culture
     - cityName: Tokyo
       aliases: Tokyo, TYO, Japan
       info: Capital city of Japan, famous for technology and tradition
   ```

2. **Knowledge is automatically indexed and retrieved during conversations**:
   - When the agent is created, knowledge is indexed into both GORM entities and sqlite-vec tables
   - During conversations, relevant knowledge is retrieved using fast vector similarity search
   - Retrieved knowledge is injected into the agent's prompt for accurate, context-aware responses

#### Technical Details

- **Vector Dimensions**: 1536 (OpenAI text-embedding-3-small)
- **Similarity Metric**: sqlite-vec distance (L2 distance)
- **Storage**: GORM entities + sqlite-vec virtual tables
- **Text Extraction**: Intelligent extraction from knowledge maps with priority field ordering
- **Performance**: High-speed native SQLite vector operations

## Installation

### Option 1: Download pre-built binaries (Recommended)

```bash
# For macOS (Apple Silicon)
curl -L https://github.com/habiliai/agentruntime/releases/latest/download/agentruntime-darwin-arm64 -o agentruntime
chmod +x agentruntime
sudo mv agentruntime /usr/local/bin/

# For macOS (Intel)
curl -L https://github.com/habiliai/agentruntime/releases/latest/download/agentruntime-darwin-amd64 -o agentruntime
chmod +x agentruntime
sudo mv agentruntime /usr/local/bin/

# For Linux
curl -L https://github.com/habiliai/agentruntime/releases/latest/download/agentruntime-linux-amd64 -o agentruntime
chmod +x agentruntime
sudo mv agentruntime /usr/local/bin/

# For Windows (using PowerShell)
Invoke-WebRequest -Uri https://github.com/habiliai/agentruntime/releases/latest/download/agentruntime-windows-amd64.exe -OutFile agentruntime.exe
```

You can also download the binaries directly from the [releases page](https://github.com/habiliai/agentruntime/releases).

### Option 2: Build from source

Prerequisites:

- Go 1.21 or higher
- Make (optional)
- Node.js 18+ (for playground)

```bash
# Clone the repository
git clone https://github.com/habiliai/agentruntime.git
cd agentruntime

# Build the project
make build

# Or build manually
go build -o bin/agentruntime ./cmd/agentruntime
```

## Quick Start

### 1. Create an Agent Configuration

Create an agent configuration file (e.g., `assistant.agent.yaml`):

```yaml
name: Alice
model: gpt-4o
tools:
  - get_weather
system: Take a deep breath and relax. Alice can help people that they are planning and executing daily tasks.
role: Assistant
bio:
  - Alice is a conversational AI model that can help you with your daily tasks.
  - Alice can check weather conditions and provide recommendations.
message_examples:
  - name: 'USER'
    text: "What's the weather like in Tokyo?"
mcpServers:
  # Optional: Add MCP servers for extended functionality
```

### 2. Set up Environment Variables (For Server Mode)

Environment variables are only used when running the agentruntime server. For programmatic usage, API keys are passed directly through options.

Create a `.env` file from the provided example:

```bash
# Copy the example environment file
cp .env.example .env

# Edit the .env file with your API keys
nano .env  # or use any text editor
```

Example `.env` file content:

```env
# Log configuration
LOG_LEVEL=debug

# LLM API Keys
OPENAI_API_KEY=your-openai-api-key
ANTHROPIC_API_KEY=your-anthropic-api-key
XAI_API_KEY=your-xai-api-key

# Tool API Keys
OPENWEATHER_API_KEY=your-openweather-api-key
```

### 3. Run the Agent

#### Start the Agent Server

```bash
# Start the agentruntime server with your agent configuration
agentruntime examples/assistant.agent.yaml

# Or run multiple agents at once
agentruntime examples/assistant.agent.yaml examples/weather_forecaster.agent.yaml

# Or specify a directory containing agent files
agentruntime examples/

# The server will start on http://localhost:3001 by default
# Use -p flag to specify a different port
agentruntime -p 8080 examples/assistant.agent.yaml
```

The agentruntime server provides:

- REST API endpoints for creating threads and sending messages
- Simple web interface for testing agents
- Thread-based conversation management
- Multi-agent support in the same server instance

#### Programmatic Usage

Use the AgentRuntime directly in your Go application:

```go
package main

import (
    "context"
    "log"

    "github.com/habiliai/agentruntime"
    "github.com/habiliai/agentruntime/entity"
    "github.com/habiliai/agentruntime/engine"
)

func main() {
    ctx := context.Background()

    // Create an agent
    agent := entity.Agent{
        Name:   "Assistant",
        Model:  "gpt-4o",
        System: "You are a helpful assistant.",
        Tools:  []string{"get_weather"},
    }

    // Initialize the runtime with API keys passed directly
    runtime, err := agentruntime.NewAgentRuntime(ctx,
        agentruntime.WithAgent(agent),
        agentruntime.WithOpenAIAPIKey("your-openai-api-key"),
        agentruntime.WithAnthropicAPIKey("your-anthropic-api-key"),
        agentruntime.WithXAIAPIKey("your-xai-api-key"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer runtime.Close()

    // Execute a request
    response, err := runtime.Run(ctx, engine.RunRequest{
        Messages: []engine.Message{
            {Role: "user", Content: "What's the weather in Tokyo?"},
        },
    }, nil)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Agent response: %s", response.Content)
}
```

## Agent Configuration

Agents are defined using YAML configuration files with the following structure:

```yaml
# Basic Information
name: AgentName # Required: Unique identifier for the agent
model: gpt-4o # Required: LLM model to use
role: Assistant # Optional: Agent's role
system: | # Optional: System prompt
  You are a helpful assistant.

# Capabilities
tools: # Optional: List of tools the agent can use
  - tool_name
  - another_tool

skills: # Optional: Custom skills/capabilities
  - name: custom_skill
    description: What this skill does

# Personality & Examples
bio: # Optional: Agent biography/description
  - Background information
  - Capabilities and expertise

message_examples: # Optional: Few-shot examples
  - name: 'USER'
    text: 'Example user message'
  - name: 'ASSISTANT'
    text: 'Example assistant response'

# Extensions
mcpServers: # Optional: MCP server configurations
  - name: server_name
    config: {}
```

## Available Tools

The runtime includes several built-in tools:

- **get_weather**: Fetch weather information for any location
- **filesystem**: Read, write, and manage files
- **memory**: Store and retrieve information across conversations
- **git**: Interact with Git repositories
- **And more**: Extend with custom tools or MCP servers

## Development

### Project Structure

```
agentruntime/
├── cmd/agentruntime/     # CLI application
├── engine/               # Core execution engine
├── entity/               # Agent and message entities
├── internal/             # Internal packages
│   ├── genkit/          # Genkit integration
│   └── tool/            # Tool management
├── playground/           # Web interface
├── examples/            # Example agent configurations
└── agentruntime.go      # Main package API
```

### Running the Playground Locally

```bash
# Install dependencies
cd playground
yarn install

# Start the development server
yarn dev

# The playground will be available at http://localhost:3000
```

### Building Docker Images

```bash
# Build the agentruntime server image
docker-compose build

# Run with infrastructure (PostgreSQL for persistence)
docker-compose -f docker-compose.infra.yaml up -d
docker-compose up
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
