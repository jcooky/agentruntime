'use client';

import { useCallback, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Send, Bot, Home } from 'lucide-react';
import { ThemeToggle } from '@/components/theme-toggle';
import { useAddMessage, useGetMessages } from '@/hooks/agentnetwork';
import { useParams, useRouter } from 'next/navigation';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

export default function ThreadPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { data: messages, isLoading: isLoadingMessages } = useGetMessages({
    threadId: parseInt(id),
  });
  const { mutate: addMessage, isPending: isAddingMessage } = useAddMessage();
  const [input, setInput] = useState('');

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      if (!input.trim() || isAddingMessage) return;

      addMessage(
        {
          thread_id: parseInt(id),
          message: input.trim(),
        },
        {
          onSuccess: () => {
            setInput('');
          },
          onError: (error) => {
            console.error('Failed to send message:', error);
          },
        },
      );
    },
    [input, isAddingMessage, addMessage, id],
  );

  return (
    <div className="flex h-screen p-4">
      <Card className="w-full max-w-4xl mx-auto flex flex-col">
        <CardHeader className="flex flex-row items-center justify-between">
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => router.push('/')}
              title="Back to threads"
            >
              <Home className="w-5 h-5" />
            </Button>
            <CardTitle>Your Assistant</CardTitle>
          </div>
          <ThemeToggle />
        </CardHeader>

        <CardContent className="flex-1 flex flex-col">
          <ScrollArea className="flex-1 pr-4">
            <div className="space-y-4">
              {isLoadingMessages ? (
                <div className="flex justify-center">
                  <div>Loading messages...</div>
                </div>
              ) : (
                messages.map((message) => (
                  <div
                    key={message.id}
                    className={`flex ${
                      message.sender === 'USER'
                        ? 'justify-end'
                        : 'justify-start'
                    }`}
                  >
                    {message.sender !== 'USER' ? (
                      <div className="flex gap-3">
                        <div className="flex flex-col items-center">
                          <div className="w-8 h-8 bg-muted rounded-full flex items-center justify-center">
                            <Bot className="w-5 h-5" />
                          </div>
                          <span className="text-xs text-muted-foreground mt-1">
                            {message.sender}
                          </span>
                        </div>
                        <div className="max-w-[80%] rounded-lg px-4 py-2 bg-muted">
                          <div className="prose prose-sm dark:prose-invert max-w-none">
                            <ReactMarkdown remarkPlugins={[remarkGfm]}>
                              {message.content}
                            </ReactMarkdown>
                          </div>
                          <p className="text-xs opacity-70 mt-1">
                            {new Date(message.created_at).toLocaleTimeString()}
                          </p>
                        </div>
                      </div>
                    ) : (
                      <div
                        className={`max-w-[80%] rounded-lg px-4 py-2 ${
                          message.sender === 'USER'
                            ? 'bg-primary text-primary-foreground'
                            : 'bg-muted'
                        }`}
                      >
                        <p className="text-sm whitespace-pre-wrap">
                          {message.content}
                        </p>
                        <p className="text-xs opacity-70 mt-1">
                          {new Date(message.created_at).toLocaleTimeString()}
                        </p>
                      </div>
                    )}
                  </div>
                ))
              )}

              {isAddingMessage && (
                <div className="flex justify-start">
                  <div className="bg-muted rounded-lg px-4 py-2">
                    <div className="flex space-x-1">
                      <div className="w-2 h-2 bg-gray-500 rounded-full animate-bounce"></div>
                      <div
                        className="w-2 h-2 bg-gray-500 rounded-full animate-bounce"
                        style={{ animationDelay: '0.1s' }}
                      ></div>
                      <div
                        className="w-2 h-2 bg-gray-500 rounded-full animate-bounce"
                        style={{ animationDelay: '0.2s' }}
                      ></div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </ScrollArea>

          <form onSubmit={handleSubmit} className="flex gap-2 mt-4">
            <Input
              value={input}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                setInput(e.target.value)
              }
              placeholder="Type your message..."
              disabled={isAddingMessage}
              className="flex-1"
            />
            <Button type="submit" disabled={isAddingMessage || !input.trim()}>
              <Send className="w-4 h-4" />
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
