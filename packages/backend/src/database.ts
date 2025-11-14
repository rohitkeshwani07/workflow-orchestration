import Database from 'better-sqlite3';
import path from 'path';
import fs from 'fs';
import { Workflow, Execution, ChatMessage } from '@workflow/shared';

const dataDir = path.join(process.cwd(), 'data');
if (!fs.existsSync(dataDir)) {
  fs.mkdirSync(dataDir, { recursive: true });
}

const dbPath = process.env.DATABASE_PATH || path.join(dataDir, 'workflows.db');
const db = new Database(dbPath);

// Initialize database schema
db.exec(`
  CREATE TABLE IF NOT EXISTS workflows (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    nodes TEXT NOT NULL,
    edges TEXT NOT NULL,
    active INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS executions (
    id TEXT PRIMARY KEY,
    workflow_id TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    error TEXT,
    context TEXT NOT NULL,
    FOREIGN KEY (workflow_id) REFERENCES workflows (id)
  );

  CREATE TABLE IF NOT EXISTS execution_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    execution_id TEXT NOT NULL,
    node_id TEXT NOT NULL,
    status TEXT NOT NULL,
    output TEXT,
    error TEXT,
    executed_at TEXT NOT NULL,
    FOREIGN KEY (execution_id) REFERENCES executions (id)
  );

  CREATE TABLE IF NOT EXISTS chat_sessions (
    id TEXT PRIMARY KEY,
    workflow_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (workflow_id) REFERENCES workflows (id)
  );

  CREATE TABLE IF NOT EXISTS chat_messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES chat_sessions (id)
  );

  CREATE INDEX IF NOT EXISTS idx_executions_workflow ON executions(workflow_id);
  CREATE INDEX IF NOT EXISTS idx_execution_logs_execution ON execution_logs(execution_id);
  CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(session_id);
`);

export class WorkflowDB {
  static createWorkflow(workflow: Workflow): void {
    const stmt = db.prepare(`
      INSERT INTO workflows (id, name, description, nodes, edges, active, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `);

    stmt.run(
      workflow.id,
      workflow.name,
      workflow.description || null,
      JSON.stringify(workflow.nodes),
      JSON.stringify(workflow.edges),
      workflow.active ? 1 : 0,
      workflow.createdAt,
      workflow.updatedAt
    );
  }

  static getWorkflow(id: string): Workflow | null {
    const stmt = db.prepare('SELECT * FROM workflows WHERE id = ?');
    const row = stmt.get(id) as any;

    if (!row) return null;

    return {
      id: row.id,
      name: row.name,
      description: row.description,
      nodes: JSON.parse(row.nodes),
      edges: JSON.parse(row.edges),
      active: row.active === 1,
      createdAt: row.created_at,
      updatedAt: row.updated_at
    };
  }

  static getAllWorkflows(): Workflow[] {
    const stmt = db.prepare('SELECT * FROM workflows ORDER BY updated_at DESC');
    const rows = stmt.all() as any[];

    return rows.map(row => ({
      id: row.id,
      name: row.name,
      description: row.description,
      nodes: JSON.parse(row.nodes),
      edges: JSON.parse(row.edges),
      active: row.active === 1,
      createdAt: row.created_at,
      updatedAt: row.updated_at
    }));
  }

  static updateWorkflow(workflow: Workflow): void {
    const stmt = db.prepare(`
      UPDATE workflows
      SET name = ?, description = ?, nodes = ?, edges = ?, active = ?, updated_at = ?
      WHERE id = ?
    `);

    stmt.run(
      workflow.name,
      workflow.description || null,
      JSON.stringify(workflow.nodes),
      JSON.stringify(workflow.edges),
      workflow.active ? 1 : 0,
      workflow.updatedAt,
      workflow.id
    );
  }

  static deleteWorkflow(id: string): void {
    db.prepare('DELETE FROM workflows WHERE id = ?').run(id);
  }

  static createExecution(execution: Execution): void {
    const stmt = db.prepare(`
      INSERT INTO executions (id, workflow_id, status, started_at, finished_at, error, context)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `);

    stmt.run(
      execution.id,
      execution.workflowId,
      execution.status,
      execution.startedAt,
      execution.finishedAt || null,
      execution.error || null,
      JSON.stringify(execution.context)
    );
  }

  static updateExecution(execution: Execution): void {
    const stmt = db.prepare(`
      UPDATE executions
      SET status = ?, finished_at = ?, error = ?, context = ?
      WHERE id = ?
    `);

    stmt.run(
      execution.status,
      execution.finishedAt || null,
      execution.error || null,
      JSON.stringify(execution.context),
      execution.id
    );
  }

  static getExecution(id: string): Execution | null {
    const stmt = db.prepare('SELECT * FROM executions WHERE id = ?');
    const row = stmt.get(id) as any;

    if (!row) return null;

    return {
      id: row.id,
      workflowId: row.workflow_id,
      status: row.status,
      startedAt: row.started_at,
      finishedAt: row.finished_at,
      error: row.error,
      context: JSON.parse(row.context)
    };
  }

  static getWorkflowExecutions(workflowId: string, limit = 50): Execution[] {
    const stmt = db.prepare(`
      SELECT * FROM executions
      WHERE workflow_id = ?
      ORDER BY started_at DESC
      LIMIT ?
    `);
    const rows = stmt.all(workflowId, limit) as any[];

    return rows.map(row => ({
      id: row.id,
      workflowId: row.workflow_id,
      status: row.status,
      startedAt: row.started_at,
      finishedAt: row.finished_at,
      error: row.error,
      context: JSON.parse(row.context)
    }));
  }

  static addExecutionLog(executionId: string, nodeId: string, status: string, output: any, error?: string): void {
    const stmt = db.prepare(`
      INSERT INTO execution_logs (execution_id, node_id, status, output, error, executed_at)
      VALUES (?, ?, ?, ?, ?, ?)
    `);

    stmt.run(
      executionId,
      nodeId,
      status,
      JSON.stringify(output),
      error || null,
      new Date().toISOString()
    );
  }

  static getExecutionLogs(executionId: string): any[] {
    const stmt = db.prepare('SELECT * FROM execution_logs WHERE execution_id = ? ORDER BY executed_at ASC');
    return stmt.all(executionId);
  }

  static createChatSession(sessionId: string, workflowId: string): void {
    const stmt = db.prepare('INSERT INTO chat_sessions (id, workflow_id, created_at) VALUES (?, ?, ?)');
    stmt.run(sessionId, workflowId, new Date().toISOString());
  }

  static addChatMessage(message: ChatMessage): void {
    const stmt = db.prepare(`
      INSERT INTO chat_messages (id, session_id, role, content, timestamp)
      VALUES (?, ?, ?, ?, ?)
    `);

    stmt.run(message.id, message.sessionId, message.role, message.content, message.timestamp);
  }

  static getChatMessages(sessionId: string): ChatMessage[] {
    const stmt = db.prepare('SELECT * FROM chat_messages WHERE session_id = ? ORDER BY timestamp ASC');
    const rows = stmt.all(sessionId) as any[];

    return rows.map(row => ({
      id: row.id,
      sessionId: row.session_id,
      role: row.role,
      content: row.content,
      timestamp: row.timestamp
    }));
  }
}

export default db;
