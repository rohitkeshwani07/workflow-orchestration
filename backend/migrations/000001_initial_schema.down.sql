-- Rollback initial schema

DROP INDEX IF EXISTS idx_chat_messages_timestamp;
DROP INDEX IF EXISTS idx_chat_messages_session_id;
DROP TABLE IF EXISTS chat_messages;

DROP INDEX IF EXISTS idx_chat_sessions_workflow_id;
DROP TABLE IF EXISTS chat_sessions;

DROP INDEX IF EXISTS idx_execution_logs_executed_at;
DROP INDEX IF EXISTS idx_execution_logs_execution_id;
DROP TABLE IF EXISTS execution_logs;

DROP INDEX IF EXISTS idx_executions_started_at;
DROP INDEX IF EXISTS idx_executions_status;
DROP INDEX IF EXISTS idx_executions_workflow_id;
DROP TABLE IF EXISTS executions;

DROP INDEX IF EXISTS idx_workflows_updated_at;
DROP INDEX IF EXISTS idx_workflows_active;
DROP TABLE IF EXISTS workflows;
