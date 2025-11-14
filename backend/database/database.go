package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/workflow-orchestration/backend/models"
)

type DB struct {
	conn *sql.DB
}

func New(dbPath string) (*DB, error) {
	// Create data directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

func (db *DB) initSchema() error {
	schema := `
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
	`

	_, err := db.conn.Exec(schema)
	return err
}

func (db *DB) Close() error {
	return db.conn.Close()
}

// Workflow operations
func (db *DB) CreateWorkflow(w *models.Workflow) error {
	nodesJSON, err := json.Marshal(w.Nodes)
	if err != nil {
		return err
	}
	edgesJSON, err := json.Marshal(w.Edges)
	if err != nil {
		return err
	}

	active := 0
	if w.Active {
		active = 1
	}

	_, err = db.conn.Exec(`
		INSERT INTO workflows (id, name, description, nodes, edges, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, w.ID, w.Name, w.Description, string(nodesJSON), string(edgesJSON), active, w.CreatedAt.Format(time.RFC3339), w.UpdatedAt.Format(time.RFC3339))

	return err
}

func (db *DB) GetWorkflow(id string) (*models.Workflow, error) {
	var w models.Workflow
	var nodesJSON, edgesJSON string
	var active int
	var createdAt, updatedAt string

	err := db.conn.QueryRow(`
		SELECT id, name, description, nodes, edges, active, created_at, updated_at
		FROM workflows WHERE id = ?
	`, id).Scan(&w.ID, &w.Name, &w.Description, &nodesJSON, &edgesJSON, &active, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(nodesJSON), &w.Nodes); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(edgesJSON), &w.Edges); err != nil {
		return nil, err
	}

	w.Active = active == 1
	w.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	w.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &w, nil
}

func (db *DB) GetAllWorkflows() ([]models.Workflow, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, description, nodes, edges, active, created_at, updated_at
		FROM workflows ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []models.Workflow
	for rows.Next() {
		var w models.Workflow
		var nodesJSON, edgesJSON string
		var active int
		var createdAt, updatedAt string

		err := rows.Scan(&w.ID, &w.Name, &w.Description, &nodesJSON, &edgesJSON, &active, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(nodesJSON), &w.Nodes); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(edgesJSON), &w.Edges); err != nil {
			return nil, err
		}

		w.Active = active == 1
		w.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		w.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		workflows = append(workflows, w)
	}

	return workflows, nil
}

func (db *DB) UpdateWorkflow(w *models.Workflow) error {
	nodesJSON, err := json.Marshal(w.Nodes)
	if err != nil {
		return err
	}
	edgesJSON, err := json.Marshal(w.Edges)
	if err != nil {
		return err
	}

	active := 0
	if w.Active {
		active = 1
	}

	_, err = db.conn.Exec(`
		UPDATE workflows
		SET name = ?, description = ?, nodes = ?, edges = ?, active = ?, updated_at = ?
		WHERE id = ?
	`, w.Name, w.Description, string(nodesJSON), string(edgesJSON), active, w.UpdatedAt.Format(time.RFC3339), w.ID)

	return err
}

func (db *DB) DeleteWorkflow(id string) error {
	_, err := db.conn.Exec("DELETE FROM workflows WHERE id = ?", id)
	return err
}

// Execution operations
func (db *DB) CreateExecution(e *models.Execution) error {
	contextJSON, err := json.Marshal(e.Context)
	if err != nil {
		return err
	}

	var finishedAt *string
	if e.FinishedAt != nil {
		t := e.FinishedAt.Format(time.RFC3339)
		finishedAt = &t
	}

	_, err = db.conn.Exec(`
		INSERT INTO executions (id, workflow_id, status, started_at, finished_at, error, context)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, e.ID, e.WorkflowID, e.Status, e.StartedAt.Format(time.RFC3339), finishedAt, e.Error, string(contextJSON))

	return err
}

func (db *DB) UpdateExecution(e *models.Execution) error {
	contextJSON, err := json.Marshal(e.Context)
	if err != nil {
		return err
	}

	var finishedAt *string
	if e.FinishedAt != nil {
		t := e.FinishedAt.Format(time.RFC3339)
		finishedAt = &t
	}

	_, err = db.conn.Exec(`
		UPDATE executions
		SET status = ?, finished_at = ?, error = ?, context = ?
		WHERE id = ?
	`, e.Status, finishedAt, e.Error, string(contextJSON), e.ID)

	return err
}

func (db *DB) GetExecution(id string) (*models.Execution, error) {
	var e models.Execution
	var contextJSON string
	var startedAt string
	var finishedAt *string

	err := db.conn.QueryRow(`
		SELECT id, workflow_id, status, started_at, finished_at, error, context
		FROM executions WHERE id = ?
	`, id).Scan(&e.ID, &e.WorkflowID, &e.Status, &startedAt, &finishedAt, &e.Error, &contextJSON)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(contextJSON), &e.Context); err != nil {
		return nil, err
	}

	e.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
	if finishedAt != nil {
		t, _ := time.Parse(time.RFC3339, *finishedAt)
		e.FinishedAt = &t
	}

	return &e, nil
}

func (db *DB) GetWorkflowExecutions(workflowID string, limit int) ([]models.Execution, error) {
	rows, err := db.conn.Query(`
		SELECT id, workflow_id, status, started_at, finished_at, error, context
		FROM executions WHERE workflow_id = ?
		ORDER BY started_at DESC LIMIT ?
	`, workflowID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []models.Execution
	for rows.Next() {
		var e models.Execution
		var contextJSON string
		var startedAt string
		var finishedAt *string

		err := rows.Scan(&e.ID, &e.WorkflowID, &e.Status, &startedAt, &finishedAt, &e.Error, &contextJSON)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(contextJSON), &e.Context); err != nil {
			return nil, err
		}

		e.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
		if finishedAt != nil {
			t, _ := time.Parse(time.RFC3339, *finishedAt)
			e.FinishedAt = &t
		}

		executions = append(executions, e)
	}

	return executions, nil
}

func (db *DB) AddExecutionLog(executionID, nodeID, status string, output *string, errMsg *string) error {
	_, err := db.conn.Exec(`
		INSERT INTO execution_logs (execution_id, node_id, status, output, error, executed_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, executionID, nodeID, status, output, errMsg, time.Now().Format(time.RFC3339))

	return err
}

func (db *DB) GetExecutionLogs(executionID string) ([]models.NodeExecutionLog, error) {
	rows, err := db.conn.Query(`
		SELECT id, execution_id, node_id, status, output, error, executed_at
		FROM execution_logs WHERE execution_id = ?
		ORDER BY executed_at ASC
	`, executionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.NodeExecutionLog
	for rows.Next() {
		var log models.NodeExecutionLog
		var executedAt string

		err := rows.Scan(&log.ID, &log.ExecutionID, &log.NodeID, &log.Status, &log.Output, &log.Error, &executedAt)
		if err != nil {
			return nil, err
		}

		log.ExecutedAt, _ = time.Parse(time.RFC3339, executedAt)
		logs = append(logs, log)
	}

	return logs, nil
}

// Chat operations
func (db *DB) CreateChatSession(sessionID, workflowID string) error {
	_, err := db.conn.Exec(`
		INSERT INTO chat_sessions (id, workflow_id, created_at)
		VALUES (?, ?, ?)
	`, sessionID, workflowID, time.Now().Format(time.RFC3339))

	return err
}

func (db *DB) AddChatMessage(msg *models.ChatMessage) error {
	_, err := db.conn.Exec(`
		INSERT INTO chat_messages (id, session_id, role, content, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`, msg.ID, msg.SessionID, msg.Role, msg.Content, msg.Timestamp.Format(time.RFC3339))

	return err
}

func (db *DB) GetChatMessages(sessionID string) ([]models.ChatMessage, error) {
	rows, err := db.conn.Query(`
		SELECT id, session_id, role, content, timestamp
		FROM chat_messages WHERE session_id = ?
		ORDER BY timestamp ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		var timestamp string

		err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &timestamp)
		if err != nil {
			return nil, err
		}

		msg.Timestamp, _ = time.Parse(time.RFC3339, timestamp)
		messages = append(messages, msg)
	}

	return messages, nil
}
