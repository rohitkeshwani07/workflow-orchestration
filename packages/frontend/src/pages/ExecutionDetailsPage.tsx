import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  CheckCircle2,
  XCircle,
  Activity,
  Clock,
  AlertCircle,
  ChevronDown,
  ChevronRight,
  Filter
} from 'lucide-react';
import { workflowApi } from '../api/workflows';
import { NodeExecutionLog, ExecutionLog } from '@workflow/shared';

export default function ExecutionDetailsPage() {
  const { id, executionId } = useParams<{ id: string; executionId: string }>();
  const navigate = useNavigate();
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  const [logLevelFilter, setLogLevelFilter] = useState<string>('all');

  const { data: workflow } = useQuery({
    queryKey: ['workflow', id],
    queryFn: async () => {
      const response = await workflowApi.getById(id!);
      return response.data.data;
    },
    enabled: !!id
  });

  const { data: nodeExecutions, isLoading: nodeLoading } = useQuery({
    queryKey: ['node-executions', executionId],
    queryFn: async () => {
      const response = await workflowApi.getNodeExecutions(executionId!);
      return response.data.data;
    },
    enabled: !!executionId,
    refetchInterval: 5000
  });

  const { data: logs, isLoading: logsLoading } = useQuery({
    queryKey: ['execution-logs', executionId, logLevelFilter],
    queryFn: async () => {
      const filters: any = { execution_id: executionId, limit: 1000 };
      if (logLevelFilter !== 'all') {
        filters.level = logLevelFilter.toUpperCase();
      }
      const response = await workflowApi.getExecutionLogs(filters);
      return response.data.data;
    },
    enabled: !!executionId,
    refetchInterval: 5000
  });

  const toggleNodeExpanded = (nodeId: string) => {
    const newExpanded = new Set(expandedNodes);
    if (newExpanded.has(nodeId)) {
      newExpanded.delete(nodeId);
    } else {
      newExpanded.add(nodeId);
    }
    setExpandedNodes(newExpanded);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
        return 'text-green-600';
      case 'error':
        return 'text-red-600';
      case 'running':
        return 'text-blue-600';
      case 'skipped':
        return 'text-gray-500';
      default:
        return 'text-gray-600';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle2 className="h-5 w-5" />;
      case 'error':
        return <XCircle className="h-5 w-5" />;
      case 'running':
        return <Activity className="h-5 w-5 animate-pulse" />;
      case 'skipped':
        return <Clock className="h-5 w-5" />;
      default:
        return <AlertCircle className="h-5 w-5" />;
    }
  };

  const getLogLevelColor = (level: string) => {
    switch (level) {
      case 'ERROR':
        return 'text-red-600 bg-red-50';
      case 'WARN':
        return 'text-yellow-600 bg-yellow-50';
      case 'INFO':
        return 'text-blue-600 bg-blue-50';
      case 'DEBUG':
        return 'text-gray-600 bg-gray-50';
      default:
        return 'text-gray-600 bg-gray-50';
    }
  };

  const formatDuration = (ms?: number) => {
    if (!ms) return '-';
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`;
    return `${(ms / 60000).toFixed(2)}m`;
  };

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleTimeString();
  };

  const isLoading = nodeLoading || logsLoading;

  if (isLoading && !nodeExecutions && !logs) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-gray-600">Loading execution details...</div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Header */}
      <div className="mb-6">
        <button
          onClick={() => navigate(`/workflows/${id}/executions`)}
          className="inline-flex items-center text-sm text-gray-500 hover:text-gray-700 mb-4"
        >
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to Executions
        </button>
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Execution Details</h1>
            {workflow && (
              <p className="text-gray-500 mt-1">
                Workflow: {workflow.name} • Execution ID: {executionId?.slice(0, 8)}...
              </p>
            )}
          </div>
        </div>
      </div>

      {/* Node Executions Timeline */}
      <div className="bg-white shadow-sm rounded-lg border border-gray-200 mb-6">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Node Executions</h2>
        </div>
        <div className="p-6">
          {nodeExecutions && nodeExecutions.length > 0 ? (
            <div className="space-y-4">
              {nodeExecutions.map((node: NodeExecutionLog) => (
                <div
                  key={node.id}
                  className="border border-gray-200 rounded-lg overflow-hidden"
                >
                  <div
                    className="flex items-center justify-between p-4 bg-gray-50 cursor-pointer hover:bg-gray-100"
                    onClick={() => toggleNodeExpanded(node.node_id)}
                  >
                    <div className="flex items-center space-x-3">
                      <div className={getStatusColor(node.status)}>
                        {getStatusIcon(node.status)}
                      </div>
                      <div>
                        <h3 className="font-medium text-gray-900">
                          {node.node_id}
                          <span className="ml-2 text-sm text-gray-500">({node.node_type})</span>
                        </h3>
                        <p className="text-sm text-gray-500">
                          Started: {formatTime(node.started_at)}
                          {node.duration_ms && ` • Duration: ${formatDuration(node.duration_ms)}`}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        node.status === 'success' ? 'bg-green-100 text-green-800' :
                        node.status === 'error' ? 'bg-red-100 text-red-800' :
                        node.status === 'running' ? 'bg-blue-100 text-blue-800' :
                        'bg-gray-100 text-gray-800'
                      }`}>
                        {node.status}
                      </span>
                      {expandedNodes.has(node.node_id) ? (
                        <ChevronDown className="h-5 w-5 text-gray-400" />
                      ) : (
                        <ChevronRight className="h-5 w-5 text-gray-400" />
                      )}
                    </div>
                  </div>

                  {expandedNodes.has(node.node_id) && (
                    <div className="p-4 bg-white border-t border-gray-200">
                      {node.error && (
                        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded">
                          <p className="text-sm font-medium text-red-800 mb-1">Error:</p>
                          <p className="text-sm text-red-700">{node.error}</p>
                        </div>
                      )}

                      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                        {/* Input */}
                        {node.input && (
                          <div>
                            <h4 className="text-sm font-medium text-gray-700 mb-2">Input:</h4>
                            <pre className="text-xs bg-gray-50 p-3 rounded border border-gray-200 overflow-x-auto">
                              {JSON.stringify(node.input, null, 2)}
                            </pre>
                          </div>
                        )}

                        {/* Output */}
                        {node.output && (
                          <div>
                            <h4 className="text-sm font-medium text-gray-700 mb-2">Output:</h4>
                            <pre className="text-xs bg-gray-50 p-3 rounded border border-gray-200 overflow-x-auto">
                              {JSON.stringify(node.output, null, 2)}
                            </pre>
                          </div>
                        )}
                      </div>

                      {/* Metadata */}
                      {node.metadata && Object.keys(node.metadata).length > 0 && (
                        <div className="mt-4">
                          <h4 className="text-sm font-medium text-gray-700 mb-2">Metadata:</h4>
                          <pre className="text-xs bg-gray-50 p-3 rounded border border-gray-200 overflow-x-auto">
                            {JSON.stringify(node.metadata, null, 2)}
                          </pre>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <p className="text-gray-500 text-center py-4">No node executions found.</p>
          )}
        </div>
      </div>

      {/* Execution Logs */}
      <div className="bg-white shadow-sm rounded-lg border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Execution Logs</h2>
          <div className="flex items-center space-x-2">
            <Filter className="h-4 w-4 text-gray-400" />
            <select
              value={logLevelFilter}
              onChange={(e) => setLogLevelFilter(e.target.value)}
              className="text-sm border-gray-300 rounded-md shadow-sm focus:border-blue-500 focus:ring-blue-500"
            >
              <option value="all">All Levels</option>
              <option value="debug">Debug</option>
              <option value="info">Info</option>
              <option value="warn">Warn</option>
              <option value="error">Error</option>
            </select>
          </div>
        </div>
        <div className="overflow-x-auto">
          {logs && logs.length > 0 ? (
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-32">
                    Time
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-20">
                    Level
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-32">
                    Node
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Message
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-24">
                    Duration
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {logs.map((log: ExecutionLog, idx: number) => (
                  <tr key={idx} className="hover:bg-gray-50">
                    <td className="px-4 py-3 whitespace-nowrap text-xs text-gray-500">
                      {formatTime(log.timestamp)}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      <span className={`px-2 py-1 text-xs font-medium rounded ${getLogLevelColor(log.level)}`}>
                        {log.level}
                      </span>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-xs text-gray-500">
                      {log.node_id || '-'}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-900">
                      {log.message}
                      {log.error && (
                        <div className="mt-1 text-xs text-red-600">{log.error}</div>
                      )}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-xs text-gray-500">
                      {formatDuration(log.duration_ms)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <div className="p-6 text-center text-gray-500">No logs found.</div>
          )}
        </div>
      </div>
    </div>
  );
}
