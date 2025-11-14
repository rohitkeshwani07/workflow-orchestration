import { useEffect, useCallback, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  addEdge,
  useNodesState,
  useEdgesState,
  Connection,
  Panel,
  Node as FlowNode,
  Edge as FlowEdge
} from 'reactflow';
import 'reactflow/dist/style.css';
import { workflowApi } from '../api/workflows';
import { Node, Edge, NodeType } from '@workflow/shared';
import CustomNode from '../components/nodes/CustomNode';
import NodeConfigPanel from '../components/NodeConfigPanel';
import NodePalette from '../components/NodePalette';
import { Save, Play, MessageSquare, ClipboardList } from 'lucide-react';

const nodeTypes = {
  [NodeType.CHAT_TRIGGER]: CustomNode,
  [NodeType.AI_AGENT]: CustomNode,
  [NodeType.HTTP_REQUEST]: CustomNode,
  [NodeType.TRANSFORM]: CustomNode,
  [NodeType.CONDITION]: CustomNode,
  [NodeType.CODE]: CustomNode,
  [NodeType.DELAY]: CustomNode,
  [NodeType.RESPONSE]: CustomNode
};

export default function WorkflowEditor() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [selectedNode, setSelectedNode] = useState<FlowNode | null>(null);
  const [workflowName, setWorkflowName] = useState('');

  const { data: workflow, isLoading } = useQuery({
    queryKey: ['workflow', id],
    queryFn: async () => {
      const response = await workflowApi.getById(id!);
      return response.data.data;
    },
    enabled: !!id
  });

  useEffect(() => {
    if (workflow) {
      setNodes(workflow.nodes as FlowNode[]);
      setEdges(workflow.edges as FlowEdge[]);
      setWorkflowName(workflow.name);
    }
  }, [workflow, setNodes, setEdges]);

  const saveMutation = useMutation({
    mutationFn: async () => {
      const updatedWorkflow = {
        name: workflowName,
        nodes: nodes as Node[],
        edges: edges as Edge[]
      };
      return workflowApi.update(id!, updatedWorkflow);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow', id] });
      alert('Workflow saved!');
    }
  });

  const executeMutation = useMutation({
    mutationFn: () => workflowApi.execute(id!),
    onSuccess: (response) => {
      alert(`Workflow executed! Execution ID: ${response.data.data.id}`);
    },
    onError: (error: any) => {
      alert(`Execution failed: ${error.response?.data?.error || error.message}`);
    }
  });

  const onConnect = useCallback(
    (connection: Connection) => setEdges((eds) => addEdge(connection, eds)),
    [setEdges]
  );

  const onNodeClick = useCallback((_: any, node: FlowNode) => {
    setSelectedNode(node);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

  const updateNodeData = useCallback(
    (nodeId: string, data: any) => {
      setNodes((nds) =>
        nds.map((node) =>
          node.id === nodeId ? { ...node, data: { ...node.data, ...data } } : node
        )
      );
    },
    [setNodes]
  );

  const addNode = useCallback(
    (type: NodeType) => {
      const newNode: FlowNode = {
        id: `node-${Date.now()}`,
        type,
        position: { x: Math.random() * 400, y: Math.random() * 400 },
        data: { label: type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase()) }
      };
      setNodes((nds) => [...nds, newNode]);
    },
    [setNodes]
  );

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-gray-600">Loading workflow...</div>
      </div>
    );
  }

  return (
    <div className="h-screen flex flex-col">
      <div className="bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
        <input
          type="text"
          value={workflowName}
          onChange={(e) => setWorkflowName(e.target.value)}
          className="text-xl font-bold border-none focus:outline-none focus:ring-2 focus:ring-blue-500 rounded px-2"
        />
        <div className="flex gap-2">
          <button
            onClick={() => navigate(`/workflows/${id}/executions`)}
            className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
          >
            <ClipboardList className="mr-2 h-4 w-4" />
            View Logs
          </button>
          <button
            onClick={() => navigate(`/chat/${id}`)}
            className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
          >
            <MessageSquare className="mr-2 h-4 w-4" />
            Test Chat
          </button>
          <button
            onClick={() => executeMutation.mutate()}
            disabled={executeMutation.isPending}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 disabled:opacity-50"
          >
            <Play className="mr-2 h-4 w-4" />
            Execute
          </button>
          <button
            onClick={() => saveMutation.mutate()}
            disabled={saveMutation.isPending}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50"
          >
            <Save className="mr-2 h-4 w-4" />
            Save
          </button>
        </div>
      </div>

      <div className="flex-1 flex">
        <NodePalette onAddNode={addNode} />

        <div className="flex-1">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            onPaneClick={onPaneClick}
            nodeTypes={nodeTypes}
            fitView
          >
            <Background />
            <Controls />
            <MiniMap />
            <Panel position="top-right" className="bg-white p-2 rounded shadow text-sm">
              <div>{nodes.length} nodes, {edges.length} connections</div>
            </Panel>
          </ReactFlow>
        </div>

        {selectedNode && (
          <NodeConfigPanel
            node={selectedNode}
            onUpdate={updateNodeData}
            onClose={() => setSelectedNode(null)}
          />
        )}
      </div>
    </div>
  );
}
