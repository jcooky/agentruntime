# AgentRuntime Playground

## Overview

The playground is a testing environment for AgentRuntime, designed to help developers understand and interact with the AgentRuntime ecosystem. It provides a web-based interface where you can create threads, send messages, and have conversations with agents.

This playground serves as an excellent starting point for first-time contributors who want to understand the project by experimenting with agent interactions in a local development environment.

## Key Features

- **Next.js-based UI**: Built with Next.js for a modern, responsive web interface
- **Agent Interaction**: Create threads and chat with agents in real-time
- **Local Development**: Designed to run entirely on your local machine
- **AgentNetwork Integration**: Uses the `agentruntime/client` library to communicate with AgentNetwork

## Prerequisites

Before running the playground, ensure you have the following services running:

1. **AgentNetwork Server**: The main AgentNetwork service must be running

   ```bash
   # From the root directory
   make run-agentnetwork
   ```

2. **Agent Runtime(s)**: At least one agent runtime (agentd) must be running

   ```bash
   # From the root directory
   make run-agentruntime
   ```

3. **Database Infrastructure**: PostgreSQL and other required services
   ```bash
   # From the root directory
   docker-compose -f docker-compose.infra.yaml up -d
   ```

## Installation

1. Navigate to the playground directory:

   ```bash
   cd playground
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

## Development

To run the playground in development mode:

```bash
npm run dev
```

The application will be available at [http://localhost:3000](http://localhost:3000).

### Development Features

- Hot reloading for instant feedback
- TypeScript support for type safety
- Integrated with shadcn/ui components for consistent UI

## Building for Production

To create a production build:

```bash
npm run build
```

To run the production build:

```bash
npm start
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
│   └── agentnetwork.ts   # AgentNetwork API hooks
└── lib/                  # Utility functions
```

## Usage Guide

1. **Create a Thread**: Click the "Create Thread" button on the main page
2. **Select Participants**: Choose which agents should participate in the conversation
3. **Start Chatting**: Send messages and receive responses from the selected agents
4. **View Threads**: Navigate between different conversation threads

## Environment Configuration

The playground connects to AgentNetwork using the following default configuration:

- AgentNetwork URL: `http://localhost:8090`

To modify this, update the configuration in `hooks/agentnetwork.ts`.

## Troubleshooting

### Common Issues

1. **Connection Refused Error**

   - Ensure AgentNetwork is running on port 8090
   - Check that at least one agent runtime is active

2. **CORS Errors**

   - The AgentNetwork server should have CORS properly configured
   - Verify you're accessing the playground from `http://localhost:3000`

3. **No Agents Available**
   - Make sure agent runtimes are registered with AgentNetwork
   - Check agent runtime logs for registration errors

## Contributing

When contributing to the playground:

1. Test your changes with both AgentNetwork and agent runtimes running
2. Ensure TypeScript types are properly defined
3. Follow the existing code style and component patterns
4. Update this README if you add new features or change setup procedures

## Learn More

- [AgentRuntime Documentation](../README.md)
- [Next.js Documentation](https://nextjs.org/docs)
- [shadcn/ui Components](https://ui.shadcn.com)
