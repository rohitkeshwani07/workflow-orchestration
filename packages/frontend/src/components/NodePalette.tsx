import { NodeType } from '@workflow/shared';
import { MessageSquare, Bot, Globe, Code, GitBranch, Timer, MessageCircle } from 'lucide-react';

interface NodePaletteProps {
  onAddNode: (type: NodeType) => void;
}

const nodeDefinitions = [
  { type: NodeType.CHAT_TRIGGER, label: 'Chat Trigger', icon: MessageSquare, color: 'purple' },
  { type: NodeType.AI_AGENT, label: 'AI Agent', icon: Bot, color: 'blue' },
  { type: NodeType.HTTP_REQUEST, label: 'HTTP Request', icon: Globe, color: 'green' },
  { type: NodeType.TRANSFORM, label: 'Transform', icon: Code, color: 'yellow' },
  { type: NodeType.CONDITION, label: 'Condition', icon: GitBranch, color: 'orange' },
  { type: NodeType.CODE, label: 'Code', icon: Code, color: 'gray' },
  { type: NodeType.DELAY, label: 'Delay', icon: Timer, color: 'pink' },
  { type: NodeType.RESPONSE, label: 'Response', icon: MessageCircle, color: 'indigo' }
];

export default function NodePalette({ onAddNode }: NodePaletteProps) {
  return (
    <div className="w-64 bg-white border-r border-gray-200 p-4 overflow-y-auto">
      <h3 className="text-sm font-semibold text-gray-900 mb-3">Node Types</h3>
      <div className="space-y-2">
        {nodeDefinitions.map(({ type, label, icon: Icon, color }) => (
          <button
            key={type}
            onClick={() => onAddNode(type)}
            className={`w-full flex items-center gap-3 px-3 py-2 rounded-md border-2 border-${color}-300 bg-${color}-50 hover:bg-${color}-100 transition-colors text-left`}
          >
            <Icon className="w-5 h-5" />
            <span className="text-sm font-medium">{label}</span>
          </button>
        ))}
      </div>
      <div className="mt-6 p-3 bg-blue-50 rounded-md">
        <p className="text-xs text-gray-600">
          Drag and drop nodes onto the canvas, then connect them to create your workflow.
        </p>
      </div>
    </div>
  );
}
