import { Node as FlowNode } from 'reactflow';
import { NodeType } from '@workflow/shared';
import { X } from 'lucide-react';

interface NodeConfigPanelProps {
  node: FlowNode;
  onUpdate: (nodeId: string, data: any) => void;
  onClose: () => void;
}

export default function NodeConfigPanel({ node, onUpdate, onClose }: NodeConfigPanelProps) {
  const renderConfig = () => {
    switch (node.type) {
      case NodeType.AI_AGENT:
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Model</label>
              <select
                value={node.data.model || 'claude-3-5-sonnet-20241022'}
                onChange={(e) => onUpdate(node.id, { model: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="claude-3-5-sonnet-20241022">Claude 3.5 Sonnet</option>
                <option value="claude-3-opus-20240229">Claude 3 Opus</option>
                <option value="claude-3-haiku-20240307">Claude 3 Haiku</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">System Prompt</label>
              <textarea
                value={node.data.systemPrompt || ''}
                onChange={(e) => onUpdate(node.id, { systemPrompt: e.target.value })}
                placeholder="You are a helpful assistant..."
                rows={6}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
              <p className="mt-1 text-xs text-gray-500">Use {'{{variableName}}'} to reference context variables</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Temperature</label>
              <input
                type="number"
                min="0"
                max="1"
                step="0.1"
                value={node.data.temperature || 0.7}
                onChange={(e) => onUpdate(node.id, { temperature: parseFloat(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Max Tokens</label>
              <input
                type="number"
                value={node.data.maxTokens || 4096}
                onChange={(e) => onUpdate(node.id, { maxTokens: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>
        );

      case NodeType.HTTP_REQUEST:
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Method</label>
              <select
                value={node.data.method || 'GET'}
                onChange={(e) => onUpdate(node.id, { method: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="DELETE">DELETE</option>
                <option value="PATCH">PATCH</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">URL</label>
              <input
                type="text"
                value={node.data.url || ''}
                onChange={(e) => onUpdate(node.id, { url: e.target.value })}
                placeholder="https://api.example.com/endpoint"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Headers (JSON)</label>
              <textarea
                value={node.data.headersJson || '{}'}
                onChange={(e) => onUpdate(node.id, { headersJson: e.target.value })}
                placeholder='{"Content-Type": "application/json"}'
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Body (JSON)</label>
              <textarea
                value={node.data.bodyJson || ''}
                onChange={(e) => onUpdate(node.id, { bodyJson: e.target.value })}
                placeholder='{"key": "value"}'
                rows={4}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
              />
            </div>
          </div>
        );

      case NodeType.TRANSFORM:
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Expression</label>
              <textarea
                value={node.data.expression || ''}
                onChange={(e) => onUpdate(node.id, { expression: e.target.value })}
                placeholder="{ result: trigger.message.toUpperCase() }"
                rows={6}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
              />
              <p className="mt-1 text-xs text-gray-500">JavaScript expression that returns a value</p>
            </div>
          </div>
        );

      case NodeType.CONDITION:
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Condition Expression</label>
              <textarea
                value={node.data.expression || ''}
                onChange={(e) => onUpdate(node.id, { expression: e.target.value })}
                placeholder="trigger.message.includes('hello')"
                rows={4}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
              />
              <p className="mt-1 text-xs text-gray-500">JavaScript expression that returns true or false</p>
            </div>
          </div>
        );

      case NodeType.CODE:
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Code</label>
              <textarea
                value={node.data.code || ''}
                onChange={(e) => onUpdate(node.id, { code: e.target.value })}
                placeholder="return { result: 'value' };"
                rows={10}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
              />
              <p className="mt-1 text-xs text-gray-500">JavaScript code with access to context variables</p>
            </div>
          </div>
        );

      case NodeType.DELAY:
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Delay (milliseconds)</label>
              <input
                type="number"
                value={node.data.delayMs || 1000}
                onChange={(e) => onUpdate(node.id, { delayMs: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>
        );

      case NodeType.RESPONSE:
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Response Message</label>
              <textarea
                value={node.data.message || ''}
                onChange={(e) => onUpdate(node.id, { message: e.target.value })}
                placeholder="Your message here..."
                rows={4}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
              <p className="mt-1 text-xs text-gray-500">Use {'{{variableName}}'} to reference context variables</p>
            </div>
          </div>
        );

      case NodeType.CHAT_TRIGGER:
        return (
          <div className="space-y-4">
            <p className="text-sm text-gray-600">
              This node triggers the workflow when a chat message is received.
              No configuration needed.
            </p>
          </div>
        );

      default:
        return <div>No configuration available for this node type.</div>;
    }
  };

  return (
    <div className="w-96 bg-white border-l border-gray-200 p-4 overflow-y-auto">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900">Configure Node</h3>
        <button
          onClick={onClose}
          className="text-gray-400 hover:text-gray-600"
        >
          <X className="w-5 h-5" />
        </button>
      </div>

      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-1">Label</label>
        <input
          type="text"
          value={node.data.label || ''}
          onChange={(e) => onUpdate(node.id, { label: e.target.value })}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
        />
      </div>

      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
        <input
          type="text"
          value={node.data.description || ''}
          onChange={(e) => onUpdate(node.id, { description: e.target.value })}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
        />
      </div>

      <hr className="my-4" />

      {renderConfig()}
    </div>
  );
}
