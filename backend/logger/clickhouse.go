package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// ClickHouseLogger handles all logging to ClickHouse
type ClickHouseLogger struct {
	conn driver.Conn
	db   string
}

// NewClickHouseLogger creates a new ClickHouse logger instance
func NewClickHouseLogger() (*ClickHouseLogger, error) {
	clickhouseURL := os.Getenv("CLICKHOUSE_URL")
	if clickhouseURL == "" {
		clickhouseURL = "http://localhost:8123"
	}

	clickhouseDB := os.Getenv("CLICKHOUSE_DB")
	if clickhouseDB == "" {
		clickhouseDB = "logs"
	}

	username := os.Getenv("CLICKHOUSE_USER")
	if username == "" {
		username = "default"
	}

	password := os.Getenv("CLICKHOUSE_PASSWORD")

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{clickhouseURL[7:]}, // Remove http:// prefix
		Auth: clickhouse.Auth{
			Database: clickhouseDB,
			Username: username,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Test connection
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	log.Printf("Connected to ClickHouse at %s", clickhouseURL)

	return &ClickHouseLogger{
		conn: conn,
		db:   clickhouseDB,
	}, nil
}

// Close closes the ClickHouse connection
func (l *ClickHouseLogger) Close() error {
	if l.conn != nil {
		return l.conn.Close()
	}
	return nil
}

// ExecutionLog represents a log entry
type ExecutionLog struct {
	ExecutionID string
	WorkflowID  string
	NodeID      string
	Timestamp   time.Time
	Level       string
	Message     string
	Metadata    map[string]interface{}
	Status      string
	Error       string
	DurationMs  uint32
}

// LogExecution writes an execution log entry to ClickHouse
func (l *ClickHouseLogger) LogExecution(ctx context.Context, log ExecutionLog) error {
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	metadataJSON, err := json.Marshal(log.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.execution_logs (
			execution_id, workflow_id, node_id, timestamp,
			level, message, metadata, status, error, duration_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, l.db)

	err = l.conn.Exec(ctx, query,
		log.ExecutionID,
		log.WorkflowID,
		log.NodeID,
		log.Timestamp,
		log.Level,
		log.Message,
		string(metadataJSON),
		log.Status,
		log.Error,
		log.DurationMs,
	)

	if err != nil {
		return fmt.Errorf("failed to insert execution log: %w", err)
	}

	return nil
}

// NodeExecution represents a node execution record
type NodeExecution struct {
	ExecutionID  string
	WorkflowID   string
	NodeID       string
	StartedAt    time.Time
	CompletedAt  time.Time
	Status       string
	NodeType     string
	Input        map[string]interface{}
	Output       map[string]interface{}
	Error        string
	DurationMs   uint32
	RetryCount   uint8
	Metadata     map[string]interface{}
}

// LogNodeExecution writes a node execution record to ClickHouse
func (l *ClickHouseLogger) LogNodeExecution(ctx context.Context, exec NodeExecution) error {
	if exec.StartedAt.IsZero() {
		exec.StartedAt = time.Now()
	}

	inputJSON, _ := json.Marshal(exec.Input)
	outputJSON, _ := json.Marshal(exec.Output)
	metadataJSON, _ := json.Marshal(exec.Metadata)

	query := fmt.Sprintf(`
		INSERT INTO %s.node_executions (
			execution_id, workflow_id, node_id, started_at, completed_at,
			status, node_type, input, output, error, duration_ms, retry_count, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, l.db)

	err := l.conn.Exec(ctx, query,
		exec.ExecutionID,
		exec.WorkflowID,
		exec.NodeID,
		exec.StartedAt,
		exec.CompletedAt,
		exec.Status,
		exec.NodeType,
		string(inputJSON),
		string(outputJSON),
		exec.Error,
		exec.DurationMs,
		exec.RetryCount,
		string(metadataJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to insert node execution: %w", err)
	}

	return nil
}

// WorkflowExecution represents a workflow execution summary
type WorkflowExecution struct {
	ExecutionID      string
	WorkflowID       string
	StartedAt        time.Time
	CompletedAt      time.Time
	Status           string
	TriggerType      string
	TriggeredBy      string
	TotalNodes       uint16
	SuccessfulNodes  uint16
	FailedNodes      uint16
	SkippedNodes     uint16
	DurationMs       uint32
	Error            string
	Metadata         map[string]interface{}
}

// LogWorkflowExecution writes a workflow execution summary to ClickHouse
func (l *ClickHouseLogger) LogWorkflowExecution(ctx context.Context, exec WorkflowExecution) error {
	if exec.StartedAt.IsZero() {
		exec.StartedAt = time.Now()
	}

	metadataJSON, _ := json.Marshal(exec.Metadata)

	query := fmt.Sprintf(`
		INSERT INTO %s.workflow_executions (
			execution_id, workflow_id, started_at, completed_at, status,
			trigger_type, triggered_by, total_nodes, successful_nodes,
			failed_nodes, skipped_nodes, duration_ms, error, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, l.db)

	err := l.conn.Exec(ctx, query,
		exec.ExecutionID,
		exec.WorkflowID,
		exec.StartedAt,
		exec.CompletedAt,
		exec.Status,
		exec.TriggerType,
		exec.TriggeredBy,
		exec.TotalNodes,
		exec.SuccessfulNodes,
		exec.FailedNodes,
		exec.SkippedNodes,
		exec.DurationMs,
		exec.Error,
		string(metadataJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to insert workflow execution: %w", err)
	}

	return nil
}

// QueryExecutionLogs retrieves execution logs with filters
func (l *ClickHouseLogger) QueryExecutionLogs(ctx context.Context, filters map[string]interface{}, limit int) ([]ExecutionLog, error) {
	query := fmt.Sprintf("SELECT execution_id, workflow_id, node_id, timestamp, level, message, metadata, status, error, duration_ms FROM %s.execution_logs WHERE 1=1", l.db)
	args := []interface{}{}

	if executionID, ok := filters["execution_id"].(string); ok && executionID != "" {
		query += " AND execution_id = ?"
		args = append(args, executionID)
	}

	if workflowID, ok := filters["workflow_id"].(string); ok && workflowID != "" {
		query += " AND workflow_id = ?"
		args = append(args, workflowID)
	}

	if level, ok := filters["level"].(string); ok && level != "" {
		query += " AND level = ?"
		args = append(args, level)
	}

	query += " ORDER BY timestamp DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := l.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query execution logs: %w", err)
	}
	defer rows.Close()

	var logs []ExecutionLog
	for rows.Next() {
		var log ExecutionLog
		var metadataStr string

		err := rows.Scan(
			&log.ExecutionID,
			&log.WorkflowID,
			&log.NodeID,
			&log.Timestamp,
			&log.Level,
			&log.Message,
			&metadataStr,
			&log.Status,
			&log.Error,
			&log.DurationMs,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan execution log: %w", err)
		}

		if metadataStr != "" {
			json.Unmarshal([]byte(metadataStr), &log.Metadata)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// QueryNodeExecutions retrieves node execution records
func (l *ClickHouseLogger) QueryNodeExecutions(ctx context.Context, executionID string, limit int) ([]NodeExecution, error) {
	query := fmt.Sprintf(`
		SELECT execution_id, workflow_id, node_id, started_at, completed_at,
		       status, node_type, input, output, error, duration_ms, retry_count, metadata
		FROM %s.node_executions
		WHERE execution_id = ?
		ORDER BY started_at DESC
		LIMIT ?
	`, l.db)

	rows, err := l.conn.Query(ctx, query, executionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query node executions: %w", err)
	}
	defer rows.Close()

	var executions []NodeExecution
	for rows.Next() {
		var exec NodeExecution
		var inputStr, outputStr, metadataStr string

		err := rows.Scan(
			&exec.ExecutionID,
			&exec.WorkflowID,
			&exec.NodeID,
			&exec.StartedAt,
			&exec.CompletedAt,
			&exec.Status,
			&exec.NodeType,
			&inputStr,
			&outputStr,
			&exec.Error,
			&exec.DurationMs,
			&exec.RetryCount,
			&metadataStr,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan node execution: %w", err)
		}

		if inputStr != "" {
			json.Unmarshal([]byte(inputStr), &exec.Input)
		}
		if outputStr != "" {
			json.Unmarshal([]byte(outputStr), &exec.Output)
		}
		if metadataStr != "" {
			json.Unmarshal([]byte(metadataStr), &exec.Metadata)
		}

		executions = append(executions, exec)
	}

	return executions, nil
}

// QueryWorkflowExecutions retrieves workflow execution summaries
func (l *ClickHouseLogger) QueryWorkflowExecutions(ctx context.Context, workflowID string, limit int) ([]WorkflowExecution, error) {
	query := fmt.Sprintf(`
		SELECT execution_id, workflow_id, started_at, completed_at, status,
		       trigger_type, triggered_by, total_nodes, successful_nodes,
		       failed_nodes, skipped_nodes, duration_ms, error, metadata
		FROM %s.workflow_executions
		WHERE workflow_id = ?
		ORDER BY started_at DESC
		LIMIT ?
	`, l.db)

	rows, err := l.conn.Query(ctx, query, workflowID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow executions: %w", err)
	}
	defer rows.Close()

	var executions []WorkflowExecution
	for rows.Next() {
		var exec WorkflowExecution
		var metadataStr string

		err := rows.Scan(
			&exec.ExecutionID,
			&exec.WorkflowID,
			&exec.StartedAt,
			&exec.CompletedAt,
			&exec.Status,
			&exec.TriggerType,
			&exec.TriggeredBy,
			&exec.TotalNodes,
			&exec.SuccessfulNodes,
			&exec.FailedNodes,
			&exec.SkippedNodes,
			&exec.DurationMs,
			&exec.Error,
			&metadataStr,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow execution: %w", err)
		}

		if metadataStr != "" {
			json.Unmarshal([]byte(metadataStr), &exec.Metadata)
		}

		executions = append(executions, exec)
	}

	return executions, nil
}
