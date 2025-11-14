import { useQuery } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Clock, CheckCircle2, XCircle, Activity } from 'lucide-react';
import { workflowApi } from '../api/workflows';
import { WorkflowExecutionLog } from '@workflow/shared';

export default function ExecutionLogsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: workflow } = useQuery({
    queryKey: ['workflow', id],
    queryFn: async () => {
      const response = await workflowApi.getById(id!);
      return response.data.data;
    },
    enabled: !!id
  });

  const { data: executions, isLoading } = useQuery({
    queryKey: ['workflow-executions', id],
    queryFn: async () => {
      const response = await workflowApi.getWorkflowExecutionHistory(id!);
      return response.data.data;
    },
    enabled: !!id,
    refetchInterval: 5000 // Refresh every 5 seconds
  });

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
        return 'bg-green-100 text-green-800';
      case 'error':
        return 'bg-red-100 text-red-800';
      case 'running':
        return 'bg-blue-100 text-blue-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle2 className="h-4 w-4" />;
      case 'error':
        return <XCircle className="h-4 w-4" />;
      case 'running':
        return <Activity className="h-4 w-4 animate-pulse" />;
      default:
        return <Clock className="h-4 w-4" />;
    }
  };

  const formatDuration = (ms?: number) => {
    if (!ms) return '-';
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`;
    return `${(ms / 60000).toFixed(2)}m`;
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleString();
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-gray-600">Loading executions...</div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-6">
        <button
          onClick={() => navigate(`/workflows/${id}`)}
          className="inline-flex items-center text-sm text-gray-500 hover:text-gray-700 mb-4"
        >
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to Workflow
        </button>
        <h1 className="text-3xl font-bold text-gray-900">
          Execution Logs
          {workflow && <span className="text-gray-500 text-xl ml-2">- {workflow.name}</span>}
        </h1>
      </div>

      {!executions || executions.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border border-gray-300">
          <Clock className="mx-auto h-12 w-12 text-gray-400" />
          <p className="mt-2 text-gray-500">No executions yet. Execute the workflow to see logs here.</p>
        </div>
      ) : (
        <div className="bg-white shadow-sm rounded-lg border border-gray-200 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Started At
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Duration
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Nodes
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Success/Failed
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {executions.map((execution: WorkflowExecutionLog) => (
                <tr
                  key={execution.execution_id}
                  className="hover:bg-gray-50 cursor-pointer"
                  onClick={() => navigate(`/workflows/${id}/executions/${execution.execution_id}`)}
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(execution.status)}`}>
                      {getStatusIcon(execution.status)}
                      <span className="ml-1 capitalize">{execution.status}</span>
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {formatDate(execution.started_at)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {formatDuration(execution.duration_ms)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {execution.total_nodes}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    <span className="text-green-600 font-medium">{execution.successful_nodes}</span>
                    {' / '}
                    <span className="text-red-600 font-medium">{execution.failed_nodes}</span>
                    {execution.skipped_nodes > 0 && (
                      <>
                        {' / '}
                        <span className="text-gray-500">{execution.skipped_nodes} skipped</span>
                      </>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-blue-600 hover:text-blue-900">
                    View Details →
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
