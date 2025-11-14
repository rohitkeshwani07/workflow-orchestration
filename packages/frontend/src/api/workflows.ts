import axios from 'axios';
import {
  Workflow,
  Execution,
  ExecutionLog,
  NodeExecutionLog,
  WorkflowExecutionLog,
  LogFilters
} from '@workflow/shared';

const api = axios.create({
  baseURL: '/api'
});

export const workflowApi = {
  getAll: () => api.get<{ data: Workflow[] }>('/workflows'),
  getById: (id: string) => api.get<{ data: Workflow }>(`/workflows/${id}`),
  create: (workflow: Partial<Workflow>) => api.post<{ data: Workflow }>('/workflows', workflow),
  update: (id: string, workflow: Partial<Workflow>) => api.put<{ data: Workflow }>(`/workflows/${id}`, workflow),
  delete: (id: string) => api.delete(`/workflows/${id}`),
  execute: (id: string, context?: any) => api.post<{ data: Execution }>(`/workflows/${id}/execute`, { context }),
  getExecutions: (id: string) => api.get<{ data: Execution[] }>(`/workflows/${id}/executions`),
  getExecutionDetails: (workflowId: string, executionId: string) =>
    api.get<{ data: { execution: Execution, logs: any[] } }>(`/workflows/${workflowId}/executions/${executionId}`),

  // ClickHouse Execution Logs APIs
  getExecutionLogs: (filters: LogFilters) =>
    api.get<{ data: ExecutionLog[] }>('/logs/executions', { params: filters }),
  getNodeExecutions: (executionId: string) =>
    api.get<{ data: NodeExecutionLog[] }>(`/logs/nodes/${executionId}`),
  getWorkflowExecutionHistory: (workflowId: string) =>
    api.get<{ data: WorkflowExecutionLog[] }>(`/logs/workflows/${workflowId}`)
};
