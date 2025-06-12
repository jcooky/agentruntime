'use client';

import { useMemo, useState } from 'react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Plus, MessageSquare, Clock, User } from 'lucide-react';
import { ThemeToggle } from '@/components/theme-toggle';
import { useRouter } from 'next/navigation';
import { formatDistance } from 'date-fns';
import {
  useCreateThread,
  useGetAgents,
  useGetThreads,
} from '@/hooks/agentruntime';
import { useToast } from '@/hooks/use-toast';

export default function Home() {
  const router = useRouter();
  const { data: threads } = useGetThreads();
  const { mutate: createThread, isPending: isCreatingThread } =
    useCreateThread();
  const { data: agents } = useGetAgents();
  const { toast } = useToast();

  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [instruction, setInstruction] = useState('');
  const [selectedParticipants, setSelectedParticipants] = useState<string[]>(
    [],
  );

  const handleParticipantToggle = (participantName: string) => {
    setSelectedParticipants((prev) => {
      if (prev.includes(participantName)) {
        return prev.filter((p) => p !== participantName);
      } else {
        return [...prev, participantName];
      }
    });
  };

  const handleCreateThread = () => {
    if (selectedParticipants.length > 0) {
      createThread(
        {
          instruction: instruction || '',
          participants: selectedParticipants,
        },
        {
          onSuccess: (data) => {
            setInstruction('');
            setSelectedParticipants([]);
            setIsCreateDialogOpen(false);
            // Navigate to the newly created thread
            router.push(`/threads/${data.id}`);
          },
          onError: (error) => {
            console.error('Failed to create thread:', error);
            toast({
              variant: 'destructive',
              title: 'Error',
              description: 'Failed to create thread. Please try again.',
            });
          },
        },
      );
    }
  };

  const formatRelativeTime = (date: Date) => {
    return formatDistance(date, new Date(), { addSuffix: true });
  };

  console.log('threads:', threads);

  return (
    <div className="container mx-auto max-w-4xl py-8 px-4">
      <div className="mb-8 flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold mb-2">Threads</h1>
          <p className="text-muted-foreground">
            Manage your conversations with AI agents
          </p>
        </div>

        <div className="flex items-center gap-4">
          <ThemeToggle />
          <Dialog
            open={isCreateDialogOpen}
            onOpenChange={setIsCreateDialogOpen}
          >
            <DialogTrigger asChild>
              <Button>
                <Plus className="w-4 h-4 mr-2" />
                New Thread
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px]">
              <DialogHeader>
                <DialogTitle>Create New Thread</DialogTitle>
                <DialogDescription>
                  Start a new conversation thread with AI agents.
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                  <Label htmlFor="instruction">Instruction (Optional)</Label>
                  <Textarea
                    id="instruction"
                    value={instruction}
                    onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                      setInstruction(e.target.value)
                    }
                    placeholder="Enter thread instruction (optional)..."
                    rows={3}
                  />
                </div>
                <div className="grid gap-2">
                  <Label>Participants</Label>
                  <div className="border rounded-md p-3 max-h-40 overflow-y-auto">
                    {agents.map((agent) => (
                      <div
                        key={agent.name}
                        className="flex items-center space-x-2 py-1"
                      >
                        <input
                          type="checkbox"
                          id={`participant-${agent.name}`}
                          checked={selectedParticipants.includes(agent.name)}
                          onChange={() => handleParticipantToggle(agent.name)}
                          className="rounded border-gray-300 text-primary focus:ring-primary"
                        />
                        <label
                          htmlFor={`participant-${agent.name}`}
                          className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer flex-1"
                        >
                          <span className="font-medium">{agent.name}</span>
                          <span className="text-muted-foreground ml-2">
                            ({agent.role})
                          </span>
                        </label>
                      </div>
                    ))}
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Select participants for the conversation (
                    {selectedParticipants.length} selected)
                  </p>
                </div>
              </div>
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setIsCreateDialogOpen(false)}
                  disabled={isCreatingThread}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleCreateThread}
                  disabled={
                    selectedParticipants.length === 0 || isCreatingThread
                  }
                >
                  {isCreatingThread ? 'Creating...' : 'Create Thread'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <div className="grid gap-4">
        {threads.length === 0 ? (
          <Card>
            <CardContent className="text-center py-12">
              <MessageSquare className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
              <p className="text-lg font-medium mb-2">No threads yet</p>
              <p className="text-muted-foreground mb-4">
                Create your first thread to start a conversation
              </p>
              <Button onClick={() => setIsCreateDialogOpen(true)}>
                <Plus className="w-4 h-4 mr-2" />
                Create Thread
              </Button>
            </CardContent>
          </Card>
        ) : (
          threads.map((thread) => (
            <Card
              key={thread.id}
              className="hover:shadow-lg transition-shadow cursor-pointer"
            >
              <CardHeader>
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <CardTitle className="text-xl">#{thread.id}</CardTitle>
                    <CardDescription className="mt-1">
                      {thread.instruction}
                    </CardDescription>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => router.push(`/threads/${thread.id}`)}
                  >
                    Open
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex gap-6 text-sm text-muted-foreground">
                  <div className="flex items-center gap-1">
                    <User className="w-4 h-4" />
                    <span>{thread.participants.length} participants</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Clock className="w-4 h-4" />
                    <span>{formatRelativeTime(thread.created_at)}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>
    </div>
  );
}
