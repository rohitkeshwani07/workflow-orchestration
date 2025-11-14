import { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { NodeType } from '@workflow/shared';
import { MessageSquare, Bot, Globe, Code, GitBranch, Timer, MessageCircle } from 'lucide-react';

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
  [NodeType.AI_AGENT]: 'bg-blue-100 border-blue-300 text-blue-900',
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
