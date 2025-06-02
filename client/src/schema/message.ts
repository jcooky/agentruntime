import { z } from 'zod';

export const Message = z.object({
  id: z.number(),
  created_at: z.date(),
  updated_at: z.date(),
  deleted_at: z.date().nullable(),
  sender: z.string(),
  content: z.string(),
  thread_id: z.number(),
});
export const MessageContent = z.object({
  text: z.string(),
  tool_calls: z.array(
    z.object({
      name: z.string(),
      arguments: z.any(),
      result: z.any(),
    }),
  ),
});
export const MessageToolCall = MessageContent.shape.tool_calls.element;

export type Message = z.infer<typeof Message>;
export type MessageContent = z.infer<typeof MessageContent>;
export type MessageToolCall = z.infer<typeof MessageToolCall>;
