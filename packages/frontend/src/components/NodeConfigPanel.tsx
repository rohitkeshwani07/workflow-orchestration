import { Node as FlowNode } from 'reactflow';
import { NodeType } from '@workflow/shared';
import { X, Plus, Trash2 } from 'lucide-react';
import { useState, useEffect } from 'react';

interface NodeConfigPanelProps {
  node: FlowNode;
  onUpdate: (nodeId: string, data: any) => void;
  onClose: () => void;
}

interface MCPTool {
  name: string;
  description: string;
  server?: string;
}

interface Credential {
  id: string;
  name: string;
  type: string;
  description?: string;
}

export default function NodeConfigPanel({ node, onUpdate, onClose }: NodeConfigPanelProps) {
  const [newToolName, setNewToolName] = useState('');
  const [newToolDesc, setNewToolDesc] = useState('');
  const [newToolServer, setNewToolServer] = useState('');
  const [credentials, setCredentials] = useState<Credential[]>([]);

  const tools: MCPTool[] = node.data.tools || [];
  const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3001';

  useEffect(() => {
    fetchCredentials();
  }, []);

  const fetchCredentials = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/credentials`);
      const result = await response.json();
      if (result.success) {
        setCredentials(result.data);
      }
    } catch (error) {
      console.error('Failed to fetch credentials:', error);
    }
  };

  const addTool = () => {
    if (!newToolName || !newToolDesc) return;

    const newTools = [
      ...tools,
      {
        name: newToolName,
        description: newToolDesc,
        server: newToolServer || undefined,
      },
    ];

    onUpdate(node.id, { tools: newTools });
    setNewToolName('');
    setNewToolDesc('');
    setNewToolServer('');
  };

  const removeTool = (index: number) => {
    const newTools = tools.filter((_, i) => i !== index);
    onUpdate(node.id, { tools: newTools });
  };

  const renderConfig = () => {
    switch (node.type) {
      case NodeType.AI_AGENT:
        const provider = node.data.provider || 'anthropic';
        const memoryType = node.data.memoryType || 'last_messages';
        const maxMemoryMessages = node.data.maxMemoryMessages || 50;

        return (
          <div className="space-y-4">
            {/* API Key Credential */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                API Key Credential
              </label>
              <select
                value={node.data.credentialId || ''}
                onChange={(e) => onUpdate(node.id, { credentialId: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="">Use environment variable (default)</option>
                {credentials
                  .filter((c) => c.type === 'api_key')
                  .map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name} {c.description && `(${c.description})`}
                    </option>
                  ))}
              </select>
              <p className="mt-1 text-xs text-gray-500">
                Select a credential or leave empty to use ANTHROPIC_API_KEY / OPENAI_API_KEY from environment
              </p>
            </div>

            {/* Provider Selection */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                AI Provider
              </label>
              <select
                value={provider}
                onChange={(e) => {
                  const newProvider = e.target.value;
                  onUpdate(node.id, {
                    provider: newProvider,
                    // Reset model to provider's default
                    model:
                      newProvider === 'openai'
                        ? 'gpt-4-turbo-preview'
                        : 'claude-3-5-sonnet-20241022',
                  });
                }}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="anthropic">Anthropic Claude</option>
                <option value="openai">OpenAI</option>
              </select>
            </div>

            {/* Model Selection */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Model</label>
              <select
                value={node.data.model || 'claude-3-5-sonnet-20241022'}
                onChange={(e) => onUpdate(node.id, { model: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              >
                {provider === 'anthropic' ? (
                  <>
                    <option value="claude-3-5-sonnet-20241022">Claude 3.5 Sonnet</option>
                    <option value="claude-3-opus-20240229">Claude 3 Opus</option>
                    <option value="claude-3-sonnet-20240229">Claude 3 Sonnet</option>
                    <option value="claude-3-haiku-20240307">Claude 3 Haiku</option>
                  </>
                ) : (
                  <>
                    <option value="gpt-4-turbo-preview">GPT-4 Turbo</option>
                    <option value="gpt-4">GPT-4</option>
                    <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
                  </>
                )}
              </select>
            </div>

            {/* System Prompt */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                System Prompt
              </label>
              <textarea
                value={node.data.systemPrompt || ''}
                onChange={(e) => onUpdate(node.id, { systemPrompt: e.target.value })}
                placeholder="You are a helpful assistant..."
                rows={6}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
              <p className="mt-1 text-xs text-gray-500">
                Use {'{{variableName}}'} to reference context variables
              </p>
            </div>

            {/* Temperature */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Temperature
              </label>
              <input
                type="number"
                min="0"
                max="1"
                step="0.1"
                value={node.data.temperature || 0.7}
                onChange={(e) =>
                  onUpdate(node.id, { temperature: parseFloat(e.target.value) })
                }
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>

            {/* Max Tokens */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Max Tokens</label>
              <input
                type="number"
                value={node.data.maxTokens || 4096}
                onChange={(e) => onUpdate(node.id, { maxTokens: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>

            {/* Memory Configuration */}
            <div className="border-t pt-4">
              <h4 className="text-sm font-medium text-gray-900 mb-3">Memory Configuration</h4>

              <div className="space-y-3">
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="memoryEnabled"
                    checked={node.data.memoryEnabled !== false}
                    onChange={(e) => onUpdate(node.id, { memoryEnabled: e.target.checked })}
                    className="h-4 w-4 text-blue-600 border-gray-300 rounded"
                  />
                  <label
                    htmlFor="memoryEnabled"
                    className="ml-2 block text-sm text-gray-700"
                  >
                    Enable conversation memory
                  </label>
                </div>

                {node.data.memoryEnabled !== false && (
                  <>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Memory Type
                      </label>
                      <select
                        value={memoryType}
                        onChange={(e) => onUpdate(node.id, { memoryType: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                      >
                        <option value="last_messages">Last N Messages</option>
                        <option value="summary" disabled>
                          Summary (Coming Soon)
                        </option>
                        <option value="full" disabled>
                          Full History (Coming Soon)
                        </option>
                      </select>
                    </div>

                    {memoryType === 'last_messages' && (
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Number of Messages
                        </label>
                        <select
                          value={maxMemoryMessages}
                          onChange={(e) =>
                            onUpdate(node.id, {
                              maxMemoryMessages: parseInt(e.target.value),
                            })
                          }
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                        >
                          <option value="5">Last 5 messages</option>
                          <option value="10">Last 10 messages</option>
                          <option value="20">Last 20 messages</option>
                          <option value="50">Last 50 messages</option>
                          <option value="100">Last 100 messages</option>
                        </select>
                        <p className="mt-1 text-xs text-gray-500">
                          Agent will remember the last {maxMemoryMessages} messages in the
                          conversation
                        </p>
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>

            {/* MCP Tools */}
            <div className="border-t pt-4">
              <h4 className="text-sm font-medium text-gray-900 mb-3">MCP Tools</h4>

              {/* List existing tools */}
              {tools.length > 0 && (
                <div className="space-y-2 mb-3">
                  {tools.map((tool, index) => (
                    <div
                      key={index}
                      className="flex items-start justify-between p-2 bg-gray-50 rounded border"
                    >
                      <div className="flex-1">
                        <div className="font-medium text-sm">{tool.name}</div>
                        <div className="text-xs text-gray-600">{tool.description}</div>
                        {tool.server && (
                          <div className="text-xs text-blue-600 mt-1">
                            Server: {tool.server}
                          </div>
                        )}
                      </div>
                      <button
                        onClick={() => removeTool(index)}
                        className="ml-2 text-red-600 hover:text-red-800"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  ))}
                </div>
              )}

              {/* Add new tool */}
              <div className="space-y-2 p-3 bg-gray-50 rounded">
                <input
                  type="text"
                  placeholder="Tool name (e.g., read_file)"
                  value={newToolName}
                  onChange={(e) => setNewToolName(e.target.value)}
                  className="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
                />
                <input
                  type="text"
                  placeholder="Description"
                  value={newToolDesc}
                  onChange={(e) => setNewToolDesc(e.target.value)}
                  className="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
                />
                <input
                  type="text"
                  placeholder="MCP server name (optional)"
                  value={newToolServer}
                  onChange={(e) => setNewToolServer(e.target.value)}
                  className="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
                />
                <button
                  onClick={addTool}
                  disabled={!newToolName || !newToolDesc}
                  className="w-full px-3 py-1 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed flex items-center justify-center"
                >
                  <Plus className="w-4 h-4 mr-1" />
                  Add Tool
                </button>
              </div>

              <p className="mt-2 text-xs text-gray-500">
                MCP tools must be configured in the AI Agent Service. Common tools: read_file,
                write_file, git_status, web_search, query_database.
              </p>
            </div>
          </div>
        );

      case NodeType.HTTP_REQUEST:
        return (
          <div className="space-y-4">
            {/* Authentication Credential */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Authentication
              </label>
              <select
                value={node.data.credentialId || ''}
                onChange={(e) => onUpdate(node.id, { credentialId: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="">No authentication</option>
                {credentials.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name} ({c.type}) {c.description && `- ${c.description}`}
                  </option>
                ))}
              </select>
              <p className="mt-1 text-xs text-gray-500">
                Selected credential will be automatically added to request headers
              </p>
            </div>

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
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Headers (JSON)
              </label>
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
              <p className="mt-1 text-xs text-gray-500">
                JavaScript expression that returns a value
              </p>
            </div>
          </div>
        );

      case NodeType.CONDITION:
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Condition Expression
              </label>
              <textarea
                value={node.data.expression || ''}
                onChange={(e) => onUpdate(node.id, { expression: e.target.value })}
                placeholder="trigger.message.includes('hello')"
                rows={4}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
              />
              <p className="mt-1 text-xs text-gray-500">
                JavaScript expression that returns true or false
              </p>
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
              <p className="mt-1 text-xs text-gray-500">
                JavaScript code with access to context variables
              </p>
            </div>
          </div>
        );

      case NodeType.DELAY:
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Delay (milliseconds)
              </label>
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
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Response Message
              </label>
              <textarea
                value={node.data.message || ''}
                onChange={(e) => onUpdate(node.id, { message: e.target.value })}
                placeholder="Your message here..."
                rows={4}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
              <p className="mt-1 text-xs text-gray-500">
                Use {'{{variableName}}'} to reference context variables
              </p>
            </div>
          </div>
        );

      case NodeType.CHAT_TRIGGER:
        return (
          <div className="space-y-4">
            <p className="text-sm text-gray-600">
              This node triggers the workflow when a chat message is received. No configuration
              needed.
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
        <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
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
