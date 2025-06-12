import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Agent, Message, Thread } from '@/lib/types';
import { z } from 'zod';

/**
 * Hook for fetching all threads
 *
 * Provides a list of all available threads using React Query for efficient caching
 * and automatic background refetching. Returns threads as an array with initial
 * empty data to prevent loading states on component mount.
 *
 * @returns {UseQueryResult} Object containing:
 *   - data: Array of thread objects
 *   - isLoading: Loading state for initial fetch
 *   - isSuccess: Success state when data is loaded
 *   - isError: Error state if the query fails
 *   - error: Error object if the query fails
 *   - refetch: Function to manually refetch threads
 *   - isFetching: Loading state for background refetches
 *
 * @example
 * ```tsx
 * function ThreadList() {
 *   const {
 *     data: threads,
 *     isLoading,
 *     isError,
 *     error,
 *     refetch
 *   } = useGetThreads();
 *
 *   if (isLoading) return <div>Loading threads...</div>;
 *   if (isError) return <div>Error: {error.message}</div>;
 *
 *   return (
 *     <div>
 *       <h2>Threads ({threads.length})</h2>
 *       {threads.map((thread) => (
 *         <div key={thread.id} className="thread-item">
 *           <h3>{thread.title}</h3>
 *           <p>Created: {new Date(thread.created_at).toLocaleDateString()}</p>
 *         </div>
 *       ))}
 *       <button onClick={() => refetch()}>Refresh Threads</button>
 *     </div>
 *   );
 * }
 * ```
 */
export function useGetThreads() {
  return useQuery({
    queryKey: ['threads'],
    queryFn: async () => {
      const threads = await fetch(
        `${process.env.NEXT_PUBLIC_AGENTRUNTIME_ENDPOINT}/threads`,
      ).then(async (res) => {
        if (!res.ok) {
          throw new Error(await res.text());
        }
        return await res.json();
      });

      console.log('threads:', threads);

      return z.array(Thread).parse(threads);
    },
    initialData: [],
    refetchOnWindowFocus: false,
  });
}

/**
 * Hook for creating new threads
 *
 * Provides a mutation function to create new threads with participants and instructions.
 * Automatically invalidates the threads query cache on successful creation to refresh the list.
 * Includes proper error handling and loading states for better user experience.
 *
 * @returns {UseMutationResult} Object containing:
 *   - mutate: Function to trigger thread creation
 *   - mutateAsync: Async version of mutate that returns a promise
 *   - isPending: Loading state during creation
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
 *         disabled={createThread.isPending}
 *       >
 *         {createThread.isPending ? 'Creating...' : 'Create Thread'}
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
      const res = await fetch(
        `${process.env.NEXT_PUBLIC_AGENTRUNTIME_ENDPOINT}/threads`,
        {
          method: 'POST',
          body: JSON.stringify(args),
        },
      ).then((res) => res.json());
      return z.object({ id: z.number() }).parse(res);
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
 * Hook for fetching all available agents
 *
 * Retrieves comprehensive information for all available agents in the network.
 * This hook provides a complete overview of agent states, capabilities, and metadata
 * without requiring specific agent names or filtering criteria. Uses React Query
 * for efficient caching and automatic background refetching.
 *
 * @returns {UseQueryResult} Object containing:
 *   - data: Array of agent objects containing runtime details
 *   - isLoading: Loading state for the query
 *   - isSuccess: Success state when data is loaded
 *   - isError: Error state if the query fails
 *   - error: Error object if the query fails
 *   - refetch: Function to manually refetch the data
 *   - isFetching: Loading state for background refetches
 *
 * @example
 * ```tsx
 * function AgentDashboard() {
 *   const {
 *     data: agents,
 *     isLoading,
 *     isError,
 *     error
 *   } = useGetAgents();
 *
 *   if (isLoading) return <div>Loading agent information...</div>;
 *   if (isError) return <div>Error: {error.message}</div>;
 *
 *   return (
 *     <div>
 *       <h2>All Agents ({agents.length})</h2>
 *       {agents.map((agent, index) => (
 *         <div key={agent.id || index} className="agent-card">
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
export function useGetAgents() {
  return useQuery({
    queryKey: ['allAgentInfo'] as const,
    queryFn: async () => {
      const agents = await fetch(
        `${process.env.NEXT_PUBLIC_AGENTRUNTIME_ENDPOINT}/agents`,
      ).then(async (res) => {
        if (!res.ok) {
          throw new Error(await res.text());
        }
        return await res.json();
      });

      console.log('agents:', agents);
      return z.array(Agent).parse(agents);
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
 * The hook fetches messages with proper error handling and loading states.
 *
 * @param {Object} params - Parameters object
 * @param {number} params.threadId - The unique identifier of the thread to fetch messages from
 *
 * @returns {UseQueryResult} Object containing:
 *   - data: Array of message objects with id, user, content, timestamps, etc.
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
 *         <div key={message.id} className={`message ${message.user}`}>
 *           <div className="sender">{message.user}</div>
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
      const messages = await fetch(
        `${process.env.NEXT_PUBLIC_AGENTRUNTIME_ENDPOINT}/threads/${threadId}/messages`,
      ).then((res) => res.json());
      return z.array(Message).parse(messages);
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
 * 'USER' as the default sender type for better user experience.
 *
 * @returns {UseMutationResult} Object containing:
 *   - mutate: Function to trigger message creation
 *   - mutateAsync: Async version of mutate that returns a promise
 *   - isPending: Loading state during message sending
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
 *       threadId: threadId,
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
 *         disabled={addMessage.isPending}
 *         required
 *       />
 *       <button
 *         type="submit"
 *         disabled={addMessage.isPending || !message.trim()}
 *       >
 *         {addMessage.isPending ? 'Sending...' : 'Send'}
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
    mutationFn: async ({
      threadId,
      message,
    }: {
      threadId: number;
      message: string;
    }) => {
      const res = await fetch(
        `${process.env.NEXT_PUBLIC_AGENTRUNTIME_ENDPOINT}/threads/${threadId}/messages`,
        {
          method: 'POST',
          body: JSON.stringify({
            message,
          }),
        },
      );

      if (!res.ok) {
        throw new Error(await res.text());
      }

      return 'ok';
    },
    onSuccess: (data, { threadId }) => {
      queryClient.invalidateQueries({
        queryKey: ['messages', { threadId }] as const,
      });
    },
    onError: (error) => {
      console.error(error);
    },
  });
}
