import {
  AgentInfo,
  AgentRuntimeInfo,
  JsonRpc,
  Message,
  MessageToolCall,
  Rpc,
  Thread,
} from './schema/rpc';
import {
  JSONRPCClient,
  JSONRPCResponse,
  TypedJSONRPCClient,
} from 'json-rpc-2.0';

/**
 * AgentNetworkClient provides a TypeScript client for interacting with the AgentNetwork RPC server.
 * This client allows you to manage agents, threads, and messages in the AgentNetwork system.
 *
 * @example
 * ```typescript
 * import { AgentNetworkClient } from '@habiliai/agentnetwork-client';
 *
 * const client = new AgentNetworkClient('http://localhost:3000/rpc');
 *
 * // Register an agent
 * await client.RegisterAgent({
 *   addr: 'http://my-agent:8080',
 *   info: [{
 *     name: 'translation-agent',
 *     description: 'A helpful translation assistant',
 *     instructions: 'Translate text between languages accurately'
 *   }]
 * });
 *
 * // Create a thread
 * const { thread_id } = await client.CreateThread({
 *   instruction: 'Help users with translations',
 *   metadata: { topic: 'translation', language: 'multilingual' }
 * });
 *
 * // Add a message
 * const { message_id } = await client.AddMessage({
 *   thread_id,
 *   sender: 'user',
 *   content: 'Please translate "Hello" to Spanish',
 *   tool_calls: []
 * });
 * ```
 */
export class AgentNetworkClient implements Rpc {
  rpcEndpoint: string;
  client: TypedJSONRPCClient<JsonRpc>;

  /**
   * Creates a new AgentNetworkClient instance.
   *
   * @param rpcEndpoint - The URL endpoint of the AgentNetwork RPC server (e.g., 'http://localhost:3000/rpc')
   *
   * @example
   * ```typescript
   * const client = new AgentNetworkClient('http://localhost:3000/rpc');
   * ```
   */
  constructor({ rpcEndpoint }: { rpcEndpoint: string }) {
    this.rpcEndpoint = rpcEndpoint;
    this.client = new JSONRPCClient(async (jsonRpcRequest) => {
      const response = await fetch(this.rpcEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(jsonRpcRequest),
      });
      if (!response.ok) {
        throw new Error(
          `HTTP error: ${response.status} ${response.statusText}`,
        );
      }
      const jsonRpcResponse = await response.json();
      this.client.receive(jsonRpcResponse as JSONRPCResponse);
    });
  }

  /**
   * Retrieves a list of threads with pagination support.
   *
   * @param args - The query parameters
   * @param args.cursor - Optional cursor for pagination (start from specific thread)
   * @param args.limit - Optional limit on number of threads to return
   * @returns Promise resolving to an object containing the thread list and optional next cursor for pagination
   *
   * @example
   * ```typescript
   * const { threads, next_cursor } = await client.GetThreads({
   *   cursor: 123,
   *   limit: 10
   * });
   *
   * console.log('Threads:', threads);
   * console.log('Next cursor:', next_cursor);
   * ```
   */
  async GetThreads(args: {
    cursor?: number;
    limit?: number;
  }): Promise<{ threads: Thread[]; next_cursor?: number }> {
    return await this.client.request(
      `habiliai-agentnetwork-v1.GetThreads`,
      args,
    );
  }

  /**
   * Adds a new message to an existing thread.
   *
   * @param args - The message details
   * @param args.thread_id - The ID of the thread to add the message to
   * @param args.sender - The name/identifier of the message sender
   * @param args.content - The text content of the message
   * @param args.tool_calls - Array of tool calls associated with this message
   * @returns Promise resolving to an object containing the new message ID
   *
   * @example
   * ```typescript
   * const { message_id } = await client.AddMessage({
   *   thread_id: 123,
   *   sender: 'user',
   *   content: 'Hello, can you help me?',
   *   tool_calls: []
   * });
   * console.log(`Message added with ID: ${message_id}`);
   * ```
   */
  async AddMessage(args: {
    thread_id: number;
    sender: string;
    content: string;
    tool_calls: MessageToolCall[];
  }): Promise<{
    message_id: number;
  }> {
    return await this.client.request(
      `habiliai-agentnetwork-v1.AddMessage`,
      args,
    );
  }

  /**
   * Checks if the specified agents are currently live/active in the network.
   *
   * @param args - The check parameters
   * @param args.names - Array of agent names to check
   * @returns Promise that resolves if all agents are live, rejects otherwise
   *
   * @example
   * ```typescript
   * try {
   *   await client.CheckLive({ names: ['translation-agent', 'weather-agent'] });
   *   console.log('All agents are live');
   * } catch (error) {
   *   console.log('Some agents are not responding');
   * }
   * ```
   */
  async CheckLive(args: { names: string[] }): Promise<void> {
    await this.client.request(`habiliai-agentnetwork-v1.CheckLive`, args);
  }

  /**
   * Creates a new conversation thread with specified instructions and metadata.
   *
   * @param args - The thread creation parameters
   * @param args.instruction - The main instruction/purpose for this thread
   * @param args.participants - Array of agent names that will participate in this thread
   * @param args.metadata - Optional key-value metadata for the thread
   * @returns Promise resolving to an object containing the new thread ID
   *
   * @example
   * ```typescript
   * const { thread_id } = await client.CreateThread({
   *   instruction: 'Help users with language translation tasks',
   *   participants: ['translation-agent', 'quality-checker-agent'],
   *   metadata: {
   *     topic: 'translation',
   *     priority: 'high',
   *     language: 'multilingual'
   *   }
   * });
   * console.log(`Thread created with ID: ${thread_id}`);
   * ```
   */
  async CreateThread(args: {
    instruction: string;
    participants: string[];
    metadata?: Record<string, string>;
  }): Promise<{ thread_id: number }> {
    return await this.client.request(
      `habiliai-agentnetwork-v1.CreateThread`,
      args,
    );
  }

  /**
   * Deregisters/removes agents from the network.
   *
   * @param args - The deregistration parameters
   * @param args.names - Array of agent names to deregister
   * @returns Promise that resolves when agents are successfully deregistered
   *
   * @example
   * ```typescript
   * await client.DeregisterAgent({
   *   names: ['old-agent', 'deprecated-agent']
   * });
   * console.log('Agents deregistered successfully');
   * ```
   */
  async DeregisterAgent(args: { names: string[] }): Promise<void> {
    await this.client.request(`habiliai-agentnetwork-v1.DeregisterAgent`, args);
  }

  /**
   * Retrieves runtime information for specified agents or all agents.
   *
   * @param args - The query parameters
   * @param args.names - Array of specific agent names to query (ignored if all=true)
   * @param args.all - If true, returns info for all agents regardless of names array
   * @returns Promise resolving to an object containing array of agent runtime information
   *
   * @example
   * ```typescript
   * // Get info for specific agents
   * const { agent_runtime_info } = await client.GetAgentRuntimeInfo({
   *   names: ['translation-agent'],
   *   all: false
   * });
   *
   * // Get info for all agents
   * const { agent_runtime_info: allAgents } = await client.GetAgentRuntimeInfo({
   *   names: [],
   *   all: true
   * });
   *
   * console.log('Agent runtime info:', agent_runtime_info);
   * ```
   */
  async GetAgentRuntimeInfo(args: {
    names?: string[];
    all?: boolean;
  }): Promise<{ agent_runtime_info: AgentRuntimeInfo[] }> {
    return await this.client.request(
      `habiliai-agentnetwork-v1.GetAgentRuntimeInfo`,
      args,
    );
  }

  /**
   * Retrieves messages from a thread with pagination support.
   *
   * @param args - The query parameters
   * @param args.thread_id - The ID of the thread to get messages from
   * @param args.order - Sort order: 'latest' for newest first, 'oldest' for oldest first
   * @param args.cursor - Optional cursor for pagination (start from specific message)
   * @param args.limit - Optional limit on number of messages to return
   * @returns Promise resolving to messages array and optional next cursor for pagination
   *
   * @example
   * ```typescript
   * // Get latest 10 messages
   * const { messages, next_cursor } = await client.GetMessages({
   *   thread_id: 123,
   *   order: 'latest',
   *   limit: 10
   * });
   *
   * // Get next page if available
   * if (next_cursor) {
   *   const nextPage = await client.GetMessages({
   *     thread_id: 123,
   *     order: 'latest',
   *     cursor: next_cursor,
   *     limit: 10
   *   });
   * }
   * ```
   */
  async GetMessages(args: {
    thread_id: number;
    order: 'latest' | 'oldest';
    cursor?: number;
    limit?: number;
  }): Promise<{
    messages: Message[] | null;
    next_cursor: number;
  }> {
    return await this.client.request(
      `habiliai-agentnetwork-v1.GetMessages`,
      args,
    );
  }

  /**
   * Gets the total number of messages in a thread.
   *
   * @param args - The query parameters
   * @param args.thread_id - The ID of the thread
   * @returns Promise resolving to an object containing the message count
   *
   * @example
   * ```typescript
   * const { num_messages } = await client.GetNumMessages({
   *   thread_id: 123
   * });
   * console.log(`Thread has ${num_messages} messages`);
   * ```
   */
  async GetNumMessages(args: {
    thread_id: number;
  }): Promise<{ num_messages: number }> {
    return await this.client.request(
      `habiliai-agentnetwork-v1.GetNumMessages`,
      args,
    );
  }

  /**
   * Retrieves detailed information about a specific thread.
   *
   * @param args - The query parameters
   * @param args.thread_id - The ID of the thread to retrieve
   * @returns Promise resolving to the complete thread information
   *
   * @example
   * ```typescript
   * const thread = await client.GetThread({
   *   thread_id: 123
   * });
   * console.log('Thread instruction:', thread.instruction);
   * console.log('Thread metadata:', thread.metadata);
   * ```
   */
  async GetThread(args: { thread_id: number }): Promise<Thread> {
    return await this.client.request(
      `habiliai-agentnetwork-v1.GetThread`,
      args,
    );
  }

  /**
   * Checks which threads have mentioned a specific agent exactly once.
   * Useful for finding threads where an agent needs to respond.
   *
   * @param args - The query parameters
   * @param args.agent_name - The name of the agent to check mentions for
   * @returns Promise resolving to an object containing array of thread IDs
   *
   * @example
   * ```typescript
   * const { thread_ids } = await client.IsMentionedOnce({
   *   agent_name: 'translation-agent'
   * });
   *
   * console.log(`Agent mentioned once in threads: ${thread_ids.join(', ')}`);
   *
   * // Process each thread where agent was mentioned
   * for (const thread_id of thread_ids) {
   *   const { messages } = await client.GetMessages({
   *     thread_id,
   *     order: 'latest',
   *     limit: 1
   *   });
   *   // Handle the mention...
   * }
   * ```
   */
  async IsMentionedOnce(args: { agent_name: string }): Promise<{
    thread_ids: number[];
  }> {
    return await this.client.request(
      `habiliai-agentnetwork-v1.IsMentionedOnce`,
      args,
    );
  }

  /**
   * Registers one or more agents with the network.
   *
   * @param args - The registration parameters
   * @param args.addr - The network address where the agent(s) can be reached
   * @param args.info - Array of agent information objects to register
   * @returns Promise that resolves when agents are successfully registered
   *
   * @example
   * ```typescript
   * await client.RegisterAgent({
   *   addr: 'http://my-agent-server:8080',
   *   info: [
   *     {
   *       name: 'translation-agent',
   *       description: 'A helpful translation assistant',
   *       instructions: 'Translate text between different languages accurately and naturally'
   *     },
   *     {
   *       name: 'weather-agent',
   *       description: 'Provides weather information',
   *       instructions: 'Help users get current weather conditions and forecasts'
   *     }
   *   ]
   * });
   * console.log('Agents registered successfully');
   * ```
   */
  async RegisterAgent(args: {
    addr: string;
    info: AgentInfo[];
  }): Promise<void> {
    await this.client.request(`habiliai-agentnetwork-v1.RegisterAgent`, args);
  }
}
