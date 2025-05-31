import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import { AgentNetworkClient, Message } from '@habiliai/agentnetwork-client';

const client = new AgentNetworkClient({
  rpcEndpoint: 'http://localhost:9080/rpc',
});

/**
 * Hook for fetching threads with infinite scroll pagination
 *
 * Provides a paginated list of threads that can be loaded incrementally.
 * Uses React Query's infinite query pattern for efficient data loading.
 *
 * @returns {UseInfiniteQueryResult} Object containing:
 *   - data: Paginated thread data with pages array
 *   - isLoading: Loading state for initial fetch
 *   - isFetchingNextPage: Loading state for loading more pages
 *   - fetchNextPage: Function to load next page
 *   - hasNextPage: Boolean indicating if more pages are available
 *   - error: Error object if the query fails
 *
 * @example
 * ```tsx
 * function ThreadList() {
 *   const {
 *     data,
 *     isLoading,
 *     isFetchingNextPage,
 *     fetchNextPage,
 *     hasNextPage,
 *     error
 *   } = useGetThreads();
 *
 *   if (isLoading) return <div>Loading threads...</div>;
 *   if (error) return <div>Error: {error.message}</div>;
 *
 *   return (
 *     <div>
 *       {data?.pages.map((page) =>
 *         page.threads.map((thread) => (
 *           <div key={thread.id}>{thread.title}</div>
 *         ))
 *       )}
 *       {hasNextPage && (
 *         <button
 *           onClick={() => fetchNextPage()}
 *           disabled={isFetchingNextPage}
 *         >
 *           {isFetchingNextPage ? 'Loading...' : 'Load More'}
 *         </button>
 *       )}
 *     </div>
 *   );
 * }
 * ```
 */
export function useGetThreads() {
  return useInfiniteQuery({
    queryKey: ['threads'],
    queryFn: async ({ pageParam: cursor = 0 }) => {
      const { threads, next_cursor: nextCursor } = await client.GetThreads({
        cursor,
        limit: 10,
      });

      return {
        threads,
        nextCursor,
      };
    },
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    initialPageParam: 0,
    initialData: {
      pages: [],
      pageParams: [],
    },
    refetchOnWindowFocus: false,
  });
}

/**
 * Hook for creating new threads
 *
 * Provides a mutation function to create new threads with participants and instructions.
 * Automatically invalidates the threads query cache on successful creation to refresh the list.
 *
 * @returns {UseMutationResult} Object containing:
 *   - mutate: Function to trigger thread creation
 *   - mutateAsync: Async version of mutate that returns a promise
 *   - isLoading: Loading state during creation
 *   - isSuccess: Success state after creation
 *   - isError: Error state if creation fails
 *   - error: Error object if creation fails
 *   - data: Created thread data on success
 *
 * @example
 * ```tsx
 * function CreateThreadForm() {
 *   const createThread = useCreateThread();
 *   const [participants, setParticipants] = useState(['user1', 'user2']);
 *   const [instruction, setInstruction] = useState('');
 *
 *   const handleSubmit = (e) => {
 *     e.preventDefault();
 *     createThread.mutate({
 *       participants,
 *       instruction
 *     }, {
 *       onSuccess: (newThread) => {
 *         console.log('Thread created:', newThread.id);
 *         // Navigate to new thread or show success message
 *       },
 *       onError: (error) => {
 *         console.error('Failed to create thread:', error);
 *       }
 *     });
 *   };
 *
 *   return (
 *     <form onSubmit={handleSubmit}>
 *       <input
 *         value={instruction}
 *         onChange={(e) => setInstruction(e.target.value)}
 *         placeholder="Thread instruction"
 *         required
 *       />
 *       <button
 *         type="submit"
 *         disabled={createThread.isLoading}
 *       >
 *         {createThread.isLoading ? 'Creating...' : 'Create Thread'}
 *       </button>
 *     </form>
 *   );
 * }
 * ```
 */
export function useCreateThread() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationKey: ['createThread'],
    mutationFn: async (args: {
      participants: string[];
      instruction: string;
    }) => {
      return await client.CreateThread(args);
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['threads'] });
    },
    onError: (error) => {
      console.error(error);
    },
  });
}

/**
 * Hook for fetching all agent runtime information
 *
 * Retrieves comprehensive runtime information for all available agents in the network.
 * This hook provides a complete overview of agent states, capabilities, and metadata
 * without requiring specific agent names or filtering criteria.
 *
 * @returns {UseQueryResult} Object containing:
 *   - data: Array of agent info objects containing runtime details
 *   - isLoading: Loading state for the query
 *   - isSuccess: Success state when data is loaded
 *   - isError: Error state if the query fails
 *   - error: Error object if the query fails
 *   - refetch: Function to manually refetch the data
 *
 * @example
 * ```tsx
 * function AgentDashboard() {
 *   const {
 *     data: agents,
 *     isLoading,
 *     isError,
 *     error
 *   } = useGetAllAgentInfo();
 *
 *   if (isLoading) return <div>Loading agent information...</div>;
 *   if (isError) return <div>Error: {error.message}</div>;
 *
 *   return (
 *     <div>
 *       <h2>All Agents ({agents.length})</h2>
 *       {agents.map((agent, index) => (
 *         <div key={index} className="agent-card">
 *           <h3>{agent.name || `Agent ${index + 1}`}</h3>
 *           <p>Status: {agent.status}</p>
 *           <p>Runtime: {agent.runtime_version}</p>
 *         </div>
 *       ))}
 *     </div>
 *   );
 * }
 * ```
 */
export function useGetAllAgentInfo() {
  return useQuery({
    queryKey: ['allAgentInfo'] as const,
    queryFn: async () => {
      const { agent_runtime_info } = await client.GetAgentRuntimeInfo({
        all: true,
      });
      console.log('agent_runtime_info', agent_runtime_info);
      return agent_runtime_info.map((info) => info.info);
    },
    initialData: [],
    refetchOnWindowFocus: false,
  });
}

/**
 * Hook for fetching messages from a specific thread
 *
 * Retrieves all messages within a specified thread, ordered from oldest to newest.
 * Uses React Query for efficient caching and automatic background refetching.
 * The hook fetches up to 200 messages per request with automatic pagination handling.
 *
 * @param {Object} params - Parameters object
 * @param {number} params.threadId - The unique identifier of the thread to fetch messages from
 *
 * @returns {UseQueryResult} Object containing:
 *   - data: Array of message objects with id, sender, content, timestamps, etc.
 *   - isLoading: Loading state for initial data fetch
 *   - isSuccess: Success state when data is loaded
 *   - isError: Error state if the query fails
 *   - error: Error object if the query fails
 *   - refetch: Function to manually refetch messages
 *   - isFetching: Loading state for background refetches
 *
 * @example
 * ```tsx
 * function MessageList({ threadId }) {
 *   const {
 *     data: messages,
 *     isLoading,
 *     isError,
 *     error,
 *     refetch
 *   } = useGetMessages({ threadId });
 *
 *   if (isLoading) return <div>Loading messages...</div>;
 *   if (isError) return <div>Error: {error.message}</div>;
 *
 *   return (
 *     <div className="message-list">
 *       {messages.map((message) => (
 *         <div key={message.id} className={`message ${message.sender}`}>
 *           <div className="sender">{message.sender}</div>
 *           <div className="content">{message.content}</div>
 *           <div className="timestamp">
 *             {new Date(message.created_at).toLocaleTimeString()}
 *           </div>
 *         </div>
 *       ))}
 *       <button onClick={() => refetch()}>Refresh Messages</button>
 *     </div>
 *   );
 * }
 * ```
 */
export function useGetMessages({ threadId }: { threadId: number }) {
  return useQuery({
    queryKey: ['messages', { threadId }] as const,
    initialData: [],
    queryFn: async ({ queryKey: [_, { threadId }] }) => {
      console.debug('threadId:', threadId);
      const { messages } = await client.GetMessages({
        thread_id: threadId,
        order: 'oldest',
        limit: 200,
      });
      return messages ?? [];
    },
    refetchOnWindowFocus: false,
    refetchInterval: 500,
    refetchIntervalInBackground: true,
  });
}

/**
 * Hook for adding new messages to a thread
 *
 * Provides a mutation function to send new messages to a specific thread.
 * Includes optimistic updates for immediate UI feedback and automatic cache
 * invalidation to keep the message list synchronized. Messages are sent with
 * 'USER' as the default sender type.
 *
 * @returns {UseMutationResult} Object containing:
 *   - mutate: Function to trigger message creation
 *   - mutateAsync: Async version of mutate that returns a promise
 *   - isLoading: Loading state during message sending
 *   - isPending: Pending state while mutation is in progress
 *   - isSuccess: Success state after message is sent
 *   - isError: Error state if sending fails
 *   - error: Error object if sending fails
 *   - data: Response data from successful message creation
 *
 * @example
 * ```tsx
 * function MessageInput({ threadId }) {
 *   const [message, setMessage] = useState('');
 *   const addMessage = useAddMessage();
 *
 *   const handleSubmit = (e) => {
 *     e.preventDefault();
 *     if (!message.trim()) return;
 *
 *     addMessage.mutate({
 *       thread_id: threadId,
 *       message: message.trim()
 *     }, {
 *       onSuccess: () => {
 *         setMessage(''); // Clear input on success
 *         console.log('Message sent successfully');
 *       },
 *       onError: (error) => {
 *         console.error('Failed to send message:', error);
 *         // Show error toast or handle error state
 *       }
 *     });
 *   };
 *
 *   return (
 *     <form onSubmit={handleSubmit} className="message-input">
 *       <input
 *         type="text"
 *         value={message}
 *         onChange={(e) => setMessage(e.target.value)}
 *         placeholder="Type your message..."
 *         disabled={addMessage.isLoading}
 *         required
 *       />
 *       <button
 *         type="submit"
 *         disabled={addMessage.isLoading || !message.trim()}
 *       >
 *         {addMessage.isLoading ? 'Sending...' : 'Send'}
 *       </button>
 *     </form>
 *   );
 * }
 * ```
 */
export function useAddMessage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationKey: ['addMessage'],
    mutationFn: async (args: { thread_id: number; message: string }) => {
      return await client.AddMessage({
        thread_id: args.thread_id,
        sender: 'USER',
        content: args.message,
        tool_calls: [],
      });
    },
    onSuccess: (data, args) => {
      queryClient.invalidateQueries({
        queryKey: ['messages', { threadId: args.thread_id }],
      });
    },
    onError: (error) => {
      console.error(error);
    },
    onMutate: (args) => {
      queryClient.setQueryData(
        ['messages', { threadId: args.thread_id }],
        (old: Message[]): Message[] => [
          ...old,
          {
            id: 0,
            sender: 'USER',
            content: args.message,
            created_at: new Date(),
            updated_at: new Date(),
            deleted_at: null,
            thread_id: args.thread_id,
          },
        ],
      );
    },
  });
}
