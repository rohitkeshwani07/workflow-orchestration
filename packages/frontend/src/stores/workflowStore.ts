import { create } from 'zustand';
import { Node, Edge, Workflow } from '@workflow/shared';

interface WorkflowStore {
  workflow: Workflow | null;
  nodes: Node[];
  edges: Edge[];
  setWorkflow: (workflow: Workflow) => void;
  setNodes: (nodes: Node[]) => void;
  setEdges: (edges: Edge[]) => void;
  updateNodeData: (nodeId: string, data: any) => void;
  reset: () => void;
}

export const useWorkflowStore = create<WorkflowStore>((set) => ({
  workflow: null,
  nodes: [],
  edges: [],
  setWorkflow: (workflow) => set({ workflow, nodes: workflow.nodes, edges: workflow.edges }),
  setNodes: (nodes) => set({ nodes }),
  setEdges: (edges) => set({ edges }),
  updateNodeData: (nodeId, data) => set((state) => ({
    nodes: state.nodes.map(node =>
      node.id === nodeId ? { ...node, data: { ...node.data, ...data } } : node
    )
  })),
  reset: () => set({ workflow: null, nodes: [], edges: [] })
}));
