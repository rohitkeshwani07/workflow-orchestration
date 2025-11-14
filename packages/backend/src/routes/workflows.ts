import { Router } from 'express';
import { WorkflowDB } from '../database';
import { Workflow, WorkflowSchema } from '@workflow/shared';
import { v4 as uuidv4 } from 'uuid';
import { WorkflowEngine } from '../engine/WorkflowEngine';

const router = Router();
const engine = new WorkflowEngine();

// Get all workflows
router.get('/', (req, res) => {
  try {
    const workflows = WorkflowDB.getAllWorkflows();
    res.json({ success: true, data: workflows });
  } catch (error: any) {
    res.status(500).json({ success: false, error: error.message });
  }
});

// Get workflow by ID
router.get('/:id', (req, res) => {
  try {
    const workflow = WorkflowDB.getWorkflow(req.params.id);
    if (!workflow) {
      return res.status(404).json({ success: false, error: 'Workflow not found' });
    }
    res.json({ success: true, data: workflow });
  } catch (error: any) {
    res.status(500).json({ success: false, error: error.message });
  }
});

// Create workflow
router.post('/', (req, res) => {
  try {
    const now = new Date().toISOString();
    const workflow: Workflow = {
      id: uuidv4(),
      name: req.body.name || 'Untitled Workflow',
      description: req.body.description,
      nodes: req.body.nodes || [],
      edges: req.body.edges || [],
      active: req.body.active || false,
      createdAt: now,
      updatedAt: now
    };

    WorkflowSchema.parse(workflow);
    WorkflowDB.createWorkflow(workflow);
    res.json({ success: true, data: workflow });
  } catch (error: any) {
    res.status(400).json({ success: false, error: error.message });
  }
});

// Update workflow
router.put('/:id', (req, res) => {
  try {
    const existing = WorkflowDB.getWorkflow(req.params.id);
    if (!existing) {
      return res.status(404).json({ success: false, error: 'Workflow not found' });
    }

    const workflow: Workflow = {
      ...existing,
      name: req.body.name || existing.name,
      description: req.body.description,
      nodes: req.body.nodes || existing.nodes,
      edges: req.body.edges || existing.edges,
      active: req.body.active !== undefined ? req.body.active : existing.active,
      updatedAt: new Date().toISOString()
    };

    WorkflowSchema.parse(workflow);
    WorkflowDB.updateWorkflow(workflow);
    res.json({ success: true, data: workflow });
  } catch (error: any) {
    res.status(400).json({ success: false, error: error.message });
  }
});

// Delete workflow
router.delete('/:id', (req, res) => {
  try {
    WorkflowDB.deleteWorkflow(req.params.id);
    res.json({ success: true });
  } catch (error: any) {
    res.status(500).json({ success: false, error: error.message });
  }
});

// Execute workflow
router.post('/:id/execute', async (req, res) => {
  try {
    const execution = await engine.executeWorkflow(req.params.id, req.body.context || {});
    res.json({ success: true, data: execution });
  } catch (error: any) {
    res.status(500).json({ success: false, error: error.message });
  }
});

// Get workflow executions
router.get('/:id/executions', (req, res) => {
  try {
    const executions = WorkflowDB.getWorkflowExecutions(req.params.id);
    res.json({ success: true, data: executions });
  } catch (error: any) {
    res.status(500).json({ success: false, error: error.message });
  }
});

// Get execution details
router.get('/:workflowId/executions/:executionId', (req, res) => {
  try {
    const execution = WorkflowDB.getExecution(req.params.executionId);
    if (!execution) {
      return res.status(404).json({ success: false, error: 'Execution not found' });
    }

    const logs = WorkflowDB.getExecutionLogs(req.params.executionId);
    res.json({ success: true, data: { execution, logs } });
  } catch (error: any) {
    res.status(500).json({ success: false, error: error.message });
  }
});

export default router;
