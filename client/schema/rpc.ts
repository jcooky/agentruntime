import { Thread } from './thread.js';
import { AgentInfo, AgentRuntimeInfo } from './agent.js';
import { Message, MessageToolCall } from './message.js';

export * from './message.js';
export * from './thread.js';
export * from './agent.js';

export type IThreadManager = {
  CreateThread(args: {
    instruction: string;
    metadata?: Record<string, string>;
  }): Promise<{
    thread_id: number;
  }>;
  GetThread(args: { thread_id: number }): Promise<Thread>;
  GetMessages(args: {
    thread_id: number;
    order: 'latest' | 'oldest';
    cursor?: number;
    limit?: number;
  }): Promise<{
    messages: Message[] | null;
    next_cursor: number;
  }>;
  AddMessage(args: {
    thread_id: number;
    sender: string;
    content: string;
    tool_calls?: MessageToolCall[] | null;
  }): Promise<{
    message_id: number;
  }>;
  GetNumMessages(args: { thread_id: number }): Promise<{
    num_messages: number;
  }>;
  IsMentionedOnce(args: { agent_name: string }): Promise<{
    thread_ids: number[];
  }>;
  GetThreads(args: { cursor?: number; limit?: number }): Promise<{
    threads: Thread[];
    next_cursor?: number;
  }>;
};

export type IAgentManager = {
  CheckLive(args: { names: string[] }): Promise<void>;
  GetAgentRuntimeInfo(args: { names?: string[]; all?: boolean }): Promise<{
    agent_runtime_info: AgentRuntimeInfo[];
  }>;
  RegisterAgent(args: { addr: string; info: AgentInfo[] }): Promise<void>;
  DeregisterAgent(args: { names: string[] }): Promise<void>;
};

export type Rpc = IThreadManager & IAgentManager;

export type JsonRpc = {
  [K in keyof Rpc as Rpc[K] extends (
    args: Parameters<Rpc[K]>[0],
  ) => ReturnType<Rpc[K]>
    ? `habiliai-agentnetwork-v1.${Extract<K, string>}`
    : never]: Rpc[K]; // only keep keys whose value is a function
};
export { Message } from './message.js';
