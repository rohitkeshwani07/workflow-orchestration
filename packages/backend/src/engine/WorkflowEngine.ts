import { Workflow, Execution, Node, Edge, NodeType } from '@workflow/shared';
import { WorkflowDB } from '../database';
import { v4 as uuidv4 } from 'uuid';
import { NodeExecutor } from './NodeExecutor';

export class WorkflowEngine {
  private nodeExecutor: NodeExecutor;

  constructor() {
    this.nodeExecutor = new NodeExecutor();
  }

  async executeWorkflow(workflowId: string, initialContext: Record<string, any> = {}): Promise<Execution> {
    const workflow = WorkflowDB.getWorkflow(workflowId);
    if (!workflow) {
      throw new Error(`Workflow ${workflowId} not found`);
    }

    const execution: Execution = {
      id: uuidv4(),
      workflowId,
      status: 'running',
      startedAt: new Date().toISOString(),
      context: initialContext
    };

    WorkflowDB.createExecution(execution);

    try {
      // Find trigger node (entry point)
      const triggerNode = workflow.nodes.find(n => n.type === NodeType.CHAT_TRIGGER);
      if (!triggerNode) {
        throw new Error('No trigger node found in workflow');
      }

      // Execute workflow starting from trigger
      await this.executeNode(triggerNode, workflow, execution);

      execution.status = 'success';
      execution.finishedAt = new Date().toISOString();
    } catch (error: any) {
      execution.status = 'error';
      execution.error = error.message;
      execution.finishedAt = new Date().toISOString();
    }

    WorkflowDB.updateExecution(execution);
    return execution;
  }

  private async executeNode(node: Node, workflow: Workflow, execution: Execution): Promise<any> {
    try {
      // Execute the node
      const result = await this.nodeExecutor.execute(node, execution.context);

      // Log the execution
      WorkflowDB.addExecutionLog(execution.id, node.id, 'success', result);

      // Update context with node output
      execution.context[node.id] = result;

      // Find next nodes
      const nextNodes = this.getNextNodes(node.id, workflow);

      // Execute next nodes
      for (const nextNode of nextNodes) {
        await this.executeNode(nextNode, workflow, execution);
      }

      return result;
    } catch (error: any) {
      WorkflowDB.addExecutionLog(execution.id, node.id, 'error', null, error.message);
      throw error;
    }
  }

  private getNextNodes(nodeId: string, workflow: Workflow): Node[] {
    const outgoingEdges = workflow.edges.filter(e => e.source === nodeId);
    return outgoingEdges
      .map(edge => workflow.nodes.find(n => n.id === edge.target))
      .filter((n): n is Node => n !== undefined);
  }

  async executeFromTrigger(workflowId: string, triggerData: any): Promise<Execution> {
    return this.executeWorkflow(workflowId, { trigger: triggerData });
  }
}
