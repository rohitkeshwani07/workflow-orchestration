import { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { NodeType } from '@workflow/shared';
import { MessageSquare, Bot, Globe, Code, GitBranch, Timer, MessageCircle, Plus, AlertTriangle, Paperclip } from 'lucide-react';

const nodeIcons: Record<NodeType, any> = {
  [NodeType.CHAT_TRIGGER]: MessageSquare,
  [NodeType.AI_AGENT]: Bot,
  [NodeType.HTTP_REQUEST]: Globe,
  [NodeType.TRANSFORM]: Code,
  [NodeType.CONDITION]: GitBranch,
  [NodeType.CODE]: Code,
  [NodeType.DELAY]: Timer,
  [NodeType.RESPONSE]: MessageCircle
};

const nodeColors: Record<NodeType, string> = {
  [NodeType.CHAT_TRIGGER]: 'bg-purple-100 border-purple-300 text-purple-900',
  [NodeType.AI_AGENT]: 'bg-gray-700 border-gray-600 text-white',
  [NodeType.HTTP_REQUEST]: 'bg-green-100 border-green-300 text-green-900',
  [NodeType.TRANSFORM]: 'bg-yellow-100 border-yellow-300 text-yellow-900',
  [NodeType.CONDITION]: 'bg-orange-100 border-orange-300 text-orange-900',
  [NodeType.CODE]: 'bg-gray-100 border-gray-300 text-gray-900',
  [NodeType.DELAY]: 'bg-pink-100 border-pink-300 text-pink-900',
  [NodeType.RESPONSE]: 'bg-indigo-100 border-indigo-300 text-indigo-900'
};

function CustomNode({ data, type, selected }: NodeProps) {
  const Icon = nodeIcons[type as NodeType] || Code;
  const colorClass = nodeColors[type as NodeType] || 'bg-gray-100 border-gray-300 text-gray-900';

  // Special rendering for AI Agent node
  if (type === NodeType.AI_AGENT) {
    const hasModelConfig = data.model || data.provider;
    const hasMemoryConfig = data.memoryEnabled !== false;
    const hasTools = data.tools && data.tools.length > 0;

    return (
      <div className="relative">
        <Handle type="target" position={Position.Top} className="w-3 h-3 !bg-gray-400" />

        {/* Main AI Agent Node */}
        <div className={`px-4 py-3 shadow-lg rounded-lg border-2 ${colorClass} ${selected ? 'ring-2 ring-blue-500' : ''} min-w-[200px]`}>
          <div className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2">
              <Bot className="w-5 h-5" />
              <div className="font-bold text-sm">AI Agent</div>
            </div>
            {(!hasModelConfig || !hasMemoryConfig || !hasTools) && (
              <AlertTriangle className="w-4 h-4 text-red-400" />
            )}
          </div>
        </div>

        {/* Attachment Sections */}
        <div className="mt-2 space-y-1">
          {/* Chat Model Attachment */}
          <div className="relative">
            <Handle
              type="source"
              position={Position.Bottom}
              id="model"
              className="w-3 h-3 !bg-purple-500"
              style={{ top: '12px', left: '50%', transform: 'translateX(-50%)' }}
            />
            <div className={`bg-gray-600 rounded px-3 py-1.5 flex items-center justify-between text-xs text-gray-200 border ${hasModelConfig ? 'border-purple-400' : 'border-gray-500'}`}>
              <span className="font-medium">
                Chat Model<span className="text-red-400">*</span>
              </span>
              <button className="w-4 h-4 rounded-full bg-gray-700 flex items-center justify-center hover:bg-gray-600">
                <Plus className="w-3 h-3" />
              </button>
            </div>
          </div>

          {/* Memory Attachment */}
          <div className="relative">
            <Handle
              type="source"
              position={Position.Bottom}
              id="memory"
              className="w-3 h-3 !bg-green-500"
              style={{ top: '12px', left: '50%', transform: 'translateX(-50%)' }}
            />
            <div className={`bg-gray-600 rounded px-3 py-1.5 flex items-center justify-between text-xs text-gray-200 border ${hasMemoryConfig ? 'border-green-400' : 'border-gray-500'}`}>
              <span className="font-medium">Memory</span>
              <button className="w-4 h-4 rounded-full bg-gray-700 flex items-center justify-center hover:bg-gray-600">
                <Plus className="w-3 h-3" />
              </button>
            </div>
          </div>

          {/* Tool Attachment */}
          <div className="relative">
            <Handle
              type="source"
              position={Position.Bottom}
              id="tool"
              className="w-3 h-3 !bg-orange-500"
              style={{ top: '12px', left: '50%', transform: 'translateX(-50%)' }}
            />
            <div className={`bg-gray-600 rounded px-3 py-1.5 flex items-center justify-between text-xs text-gray-200 border ${hasTools ? 'border-orange-400' : 'border-gray-500'}`}>
              <span className="font-medium">Tool</span>
              <button className="w-4 h-4 rounded-full bg-gray-700 flex items-center justify-center hover:bg-gray-600">
                <Plus className="w-3 h-3" />
              </button>
            </div>
          </div>
        </div>

        {/* Tools Section - MCP Client */}
        {hasTools && (
          <div className="mt-3 flex flex-col items-center">
            <div className="text-xs text-gray-400 mb-2">Tools</div>
            <div className="relative">
              <div className="w-16 h-16 rounded-full bg-gray-700 border-2 border-gray-600 flex items-center justify-center shadow-lg">
                <Paperclip className="w-8 h-8 text-gray-300" />
              </div>
              {!hasTools && (
                <div className="absolute -top-1 -right-1">
                  <AlertTriangle className="w-5 h-5 text-red-500" />
                </div>
              )}
              <div className="text-xs text-gray-300 text-center mt-2 font-medium">MCP Client</div>
            </div>
          </div>
        )}

        {/* Connector line from Tool to MCP Client */}
        {hasTools && (
          <svg className="absolute left-1/2 -translate-x-1/2" style={{ top: '205px', width: '2px', height: '40px' }}>
            <line x1="1" y1="0" x2="1" y2="40" stroke="#4B5563" strokeWidth="2" strokeDasharray="4,4" />
          </svg>
        )}

        <Handle type="source" position={Position.Right} className="w-3 h-3 !bg-gray-400" />
      </div>
    );
  }

  // Default rendering for other node types
  return (
    <div className={`px-4 py-2 shadow-md rounded-md border-2 ${colorClass} ${selected ? 'ring-2 ring-blue-500' : ''}`}>
      <Handle type="target" position={Position.Top} className="w-2 h-2" />
      <div className="flex items-center gap-2">
        <Icon className="w-4 h-4" />
        <div>
          <div className="font-bold text-sm">{data.label || type}</div>
          {data.description && (
            <div className="text-xs opacity-75">{data.description}</div>
          )}
        </div>
      </div>
      <Handle type="source" position={Position.Bottom} className="w-2 h-2" />
    </div>
  );
}

export default memo(CustomNode);
