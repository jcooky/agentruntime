# AgentRuntime Playground

## Overview

The playground is a testing environment for AgentRuntime, designed to help developers understand and interact with the agentruntime ecosystem. It provides a web-based interface where you can create threads, send messages, and have conversations with agents running on agentruntime.

This playground serves as an excellent starting point for first-time contributors who want to understand the project by experimenting with agent interactions in a local development environment.

## Key Features

- **Next.js-based UI**: Built with Next.js for a modern, responsive web interface
- **Agent Interaction**: Create threads and chat with agents in real-time
- **Local Development**: Designed to run entirely on your local machine
- **AgentRuntime Integration**: Direct communication with agentruntime servers

## Prerequisites

Before running the playground, ensure you have the following:

1. **AgentRuntime Server**: At least one agentruntime server must be running

   ```bash
   # From the root directory, run agentruntime with your agent files
   ./agentruntime path/to/your/agent/files -p 3001
   ```

   Or using Go:

   ```bash
   go run . path/to/your/agent/files -p 3001
   ```

2. **Agent Configuration**: Prepare your agent YAML configuration files

   Example agent configuration:

   ```yaml
   name: 'example-agent'
   description: 'An example agent for testing'
   model: 'claude-3-5-sonnet-20241022'
   instructions: 'You are a helpful assistant.'
   ```

## Installation

1. Navigate to the playground directory:

   ```bash
   cd playground
   ```

2. Install dependencies:
   ```bash
   yarn install
   ```

## Development

To run the playground in development mode:

```bash
yarn dev
```

The application will be available at [http://localhost:3000](http://localhost:3000).

### Development Features

- Hot reloading for instant feedback
- TypeScript support for type safety
- Integrated with shadcn/ui components for consistent UI

## Building for Production

To create a production build:

```bash
yarn build
```

To run the production build:

```bash
yarn start
```

## Project Structure

```
playground/
├── app/                    # Next.js app directory
│   ├── page.tsx           # Main page - thread list
│   └── threads/[id]/      # Thread detail page
├── components/            # React components
│   ├── ui/               # shadcn/ui components
│   └── layout.tsx        # Layout components
├── hooks/                # Custom React hooks
│   └── agentruntime.ts   # AgentRuntime API hooks
└── lib/                  # Utility functions
```

## Usage Guide

1. **Start AgentRuntime**: Ensure your agentruntime server is running with your agent configurations
2. **Create a Thread**: Click the "Create Thread" button on the main page
3. **Select Participants**: Choose which agents should participate in the conversation
4. **Start Chatting**: Send messages and receive responses from the selected agents
5. **View Threads**: Navigate between different conversation threads

## Environment Configuration

The playground connects to agentruntime using the following default configuration:

- AgentRuntime URL: `http://localhost:3001`

To modify this, update the configuration in `hooks/agentruntime.ts`.

## Troubleshooting

### Common Issues

1. **Connection Refused Error**

   - Ensure agentruntime is running on the expected port
   - Check that your agent configuration files are valid
   - Verify the agentruntime server started successfully

2. **CORS Errors**

   - The agentruntime server should have CORS properly configured
   - Verify you're accessing the playground from `http://localhost:3000`

3. **No Agents Available**

   - Make sure you have valid agent YAML configuration files
   - Check that agentruntime loaded your agents successfully
   - Review agentruntime server logs for configuration errors

4. **Agent Not Responding**
   - Verify your agent configuration includes required fields (name, model, instructions)
   - Check that API keys are properly configured for your model provider
   - Review agentruntime logs for API communication errors

## Contributing

When contributing to the playground:

1. Test your changes with agentruntime running locally
2. Ensure TypeScript types are properly defined
3. Follow the existing code style and component patterns
4. Update this README if you add new features or change setup procedures

## Learn More

- [AgentRuntime Documentation](../README.md)
- [Next.js Documentation](https://nextjs.org/docs)
- [shadcn/ui Components](https://ui.shadcn.com)
