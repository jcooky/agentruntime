import { z } from 'zod';

export const Message = z.object({
  id: z.number(),
  user: z.string(),
  content: z.string(),
  created_at: z.coerce.date(),
  updated_at: z.coerce.date(),
  thread_id: z.number(),
});

export type Message = z.infer<typeof Message>;

export const Thread = z.object({
  id: z.number(),
  instruction: z.string(),
  participants: z.array(z.string()),
  created_at: z.coerce.date(),
  updated_at: z.coerce.date(),
});

export type Thread = z.infer<typeof Thread>;

export const Agent = z.object({
  name: z.string(),
  description: z.string(),
  role: z.string(),
});

export type Agent = z.infer<typeof Agent>;
