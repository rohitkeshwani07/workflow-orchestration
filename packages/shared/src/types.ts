import { z } from 'zod';

// Node Types
export enum NodeType {
  CHAT_TRIGGER = 'chat_trigger',
  AI_AGENT = 'ai_agent',
  HTTP_REQUEST = 'http_request',
  TRANSFORM = 'transform',
  CONDITION = 'condition',
  CODE = 'code',
  DELAY = 'delay',
  RESPONSE = 'response'
}

// Base Node Schema
export const NodeSchema = z.object({
  id: z.string(),
  type: z.nativeEnum(NodeType),
  position: z.object({
    x: z.number(),
    y: z.number()
  }),
  data: z.record(z.any())
});

// Edge Schema
export const EdgeSchema = z.object({
  id: z.string(),
  source: z.string(),
  target: z.string(),
  sourceHandle: z.string().optional(),
  targetHandle: z.string().optional()
});

// Workflow Schema
export const WorkflowSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().optional(),
  nodes: z.array(NodeSchema),
  edges: z.array(EdgeSchema),
  active: z.boolean().default(false),
  createdAt: z.string(),
  updatedAt: z.string()
});

// Execution Schema
export const ExecutionSchema = z.object({
  id: z.string(),
  workflowId: z.string(),
  status: z.enum(['running', 'success', 'error', 'waiting']),
  startedAt: z.string(),
  finishedAt: z.string().optional(),
  error: z.string().optional(),
  context: z.record(z.any()).default({})
});

// Node Execution Result
export const NodeExecutionResultSchema = z.object({
  nodeId: z.string(),
  status: z.enum(['success', 'error', 'skipped']),
  output: z.any(),
  error: z.string().optional(),
  executedAt: z.string()
});

// Types
export type Node = z.infer<typeof NodeSchema>;
export type Edge = z.infer<typeof EdgeSchema>;
export type Workflow = z.infer<typeof WorkflowSchema>;
export type Execution = z.infer<typeof ExecutionSchema>;
export type NodeExecutionResult = z.infer<typeof NodeExecutionResultSchema>;

// AI Agent Configuration
export interface AIAgentConfig {
  model: string;
  systemPrompt: string;
  temperature: number;
  maxTokens: number;
  apiKey?: string;
}

// HTTP Request Configuration
export interface HTTPRequestConfig {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  url: string;
  headers?: Record<string, string>;
  body?: any;
}

// Transform Configuration
export interface TransformConfig {
  expression: string; // JavaScript expression
}

// Condition Configuration
export interface ConditionConfig {
  expression: string; // JavaScript boolean expression
}

// Chat Message
export interface ChatMessage {
  id: string;
  sessionId: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
}

// API Response Types
export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
}
