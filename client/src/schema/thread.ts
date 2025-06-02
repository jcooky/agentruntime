import { z } from 'zod';
import { Message } from './message.js';

export const Thread = z.object({
  id: z.number(),
  created_at: z.date(),
  updated_at: z.date(),
  instruction: z.string(),
  participants: z.array(z.string()),
});
export type Thread = z.infer<typeof Thread>;
