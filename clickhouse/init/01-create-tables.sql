-- Execution Logs Table
-- This table stores all workflow execution logs in an append-only fashion
-- Optimized for time-series queries with ORDER BY timestamp

CREATE TABLE IF NOT EXISTS logs.execution_logs (
    -- Identifiers
    id UUID DEFAULT generateUUIDv4(),
    execution_id String NOT NULL,
    workflow_id String NOT NULL,
    node_id String,

    -- Timestamps
    timestamp DateTime64(3) DEFAULT now64(3),

    -- Log data
    level String NOT NULL,  -- INFO, DEBUG, WARN, ERROR
    message String NOT NULL,

    -- Context data (stored as JSON strings for flexibility)
    metadata String DEFAULT '{}',

    -- Status information
    status String,  -- PENDING, RUNNING, SUCCESS, FAILED, CANCELLED
    error String DEFAULT '',

    -- Performance metrics
    duration_ms UInt32 DEFAULT 0,

    -- Index for efficient queries
    INDEX idx_execution_id execution_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_workflow_id workflow_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_level level TYPE set(10) GRANULARITY 1,
    INDEX idx_status status TYPE set(10) GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (timestamp, execution_id)
PARTITION BY toYYYYMM(timestamp)
TTL toDateTime(timestamp) + INTERVAL 90 DAY  -- Keep logs for 90 days
SETTINGS index_granularity = 8192;

-- Node Execution Table
-- Tracks individual node executions within workflows
CREATE TABLE IF NOT EXISTS logs.node_executions (
    -- Identifiers
    id UUID DEFAULT generateUUIDv4(),
    execution_id String NOT NULL,
    workflow_id String NOT NULL,
    node_id String NOT NULL,

    -- Timestamps
    started_at DateTime64(3) NOT NULL,
    completed_at DateTime64(3) DEFAULT toDateTime64(0, 3),

    -- Execution details
    status String NOT NULL,  -- PENDING, RUNNING, SUCCESS, FAILED, SKIPPED
    node_type String NOT NULL,

    -- Input/Output (stored as JSON)
    input String DEFAULT '{}',
    output String DEFAULT '{}',
    error String DEFAULT '',

    -- Performance
    duration_ms UInt32 DEFAULT 0,
    retry_count UInt8 DEFAULT 0,

    -- Metadata
    metadata String DEFAULT '{}',

    -- Indexes
    INDEX idx_execution_id execution_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_workflow_id workflow_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_node_id node_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_status status TYPE set(10) GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (started_at, execution_id, node_id)
PARTITION BY toYYYYMM(started_at)
TTL toDateTime(started_at) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- Workflow Execution Summary Table
-- High-level workflow execution tracking
CREATE TABLE IF NOT EXISTS logs.workflow_executions (
    -- Identifiers
    execution_id String NOT NULL,
    workflow_id String NOT NULL,

    -- Timestamps
    started_at DateTime64(3) NOT NULL,
    completed_at DateTime64(3) DEFAULT toDateTime64(0, 3),

    -- Status
    status String NOT NULL,  -- PENDING, RUNNING, SUCCESS, FAILED, CANCELLED

    -- Execution details
    trigger_type String DEFAULT 'manual',  -- manual, scheduled, api, webhook
    triggered_by String DEFAULT '',

    -- Results
    total_nodes UInt16 DEFAULT 0,
    successful_nodes UInt16 DEFAULT 0,
    failed_nodes UInt16 DEFAULT 0,
    skipped_nodes UInt16 DEFAULT 0,

    -- Performance
    duration_ms UInt32 DEFAULT 0,

    -- Error information
    error String DEFAULT '',

    -- Metadata
    metadata String DEFAULT '{}',

    -- Indexes
    INDEX idx_execution_id execution_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_workflow_id workflow_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_status status TYPE set(10) GRANULARITY 1,
    INDEX idx_trigger_type trigger_type TYPE set(10) GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (started_at, execution_id)
PARTITION BY toYYYYMM(started_at)
TTL toDateTime(started_at) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- Materialized view for recent execution statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS logs.execution_stats_hourly
ENGINE = SummingMergeTree()
ORDER BY (hour, workflow_id, status)
PARTITION BY toYYYYMM(hour)
TTL hour + INTERVAL 180 DAY
AS SELECT
    toStartOfHour(started_at) AS hour,
    workflow_id,
    status,
    count() AS execution_count,
    sum(duration_ms) AS total_duration_ms,
    avg(duration_ms) AS avg_duration_ms,
    max(duration_ms) AS max_duration_ms,
    min(duration_ms) AS min_duration_ms
FROM logs.workflow_executions
GROUP BY hour, workflow_id, status;
