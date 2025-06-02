import { z } from 'zod';

export const AgentRuntimeInfo = z.object({
  info: z.object({
    name: z.string(),
    role: z.string(),
    metadata: z.record(z.string()).optional(),
  }),
  addr: z.string(),
});
export const AgentInfo = AgentRuntimeInfo.shape.info;

export type AgentInfo = z.infer<typeof AgentInfo>;
export type AgentRuntimeInfo = z.infer<typeof AgentRuntimeInfo>;
