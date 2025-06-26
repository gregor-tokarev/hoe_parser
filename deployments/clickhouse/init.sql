-- Create database
CREATE DATABASE IF NOT EXISTS hoe_parser;

-- Use the database
USE hoe_parser;

-- Create events table for storing parsed events
CREATE TABLE IF NOT EXISTS events (
    id UUID DEFAULT generateUUIDv4(),
    timestamp DateTime64(3) DEFAULT now(),
    event_type String,
    source String,
    raw_data String,
    parsed_data String,
    metadata Map(String, String),
    processing_time_ms UInt32,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (timestamp, event_type, source)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 90 DAY;

-- Create metrics table for monitoring
CREATE TABLE IF NOT EXISTS metrics (
    id UUID DEFAULT generateUUIDv4(),
    timestamp DateTime64(3) DEFAULT now(),
    metric_name String,
    metric_value Float64,
    labels Map(String, String),
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (timestamp, metric_name)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 30 DAY;

-- Create parsing_errors table for error tracking
CREATE TABLE IF NOT EXISTS parsing_errors (
    id UUID DEFAULT generateUUIDv4(),
    timestamp DateTime64(3) DEFAULT now(),
    error_type String,
    error_message String,
    input_data String,
    stack_trace String,
    source String,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (timestamp, error_type, source)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 30 DAY;

-- Create materialized view for real-time metrics
CREATE MATERIALIZED VIEW IF NOT EXISTS events_summary_mv
TO metrics
AS SELECT
    now() as timestamp,
    'events_per_minute' as metric_name,
    count() as metric_value,
    map('event_type', event_type, 'source', source) as labels,
    now() as created_at
FROM events
WHERE timestamp >= now() - INTERVAL 1 MINUTE
GROUP BY event_type, source;

-- Create some sample data (optional)
INSERT INTO events (event_type, source, raw_data, parsed_data, metadata, processing_time_ms) VALUES
    ('user_action', 'web', '{"user_id": "123", "action": "click"}', '{"user_id": "123", "action": "click", "processed": true}', map('version', '1.0', 'region', 'us-east-1'), 15),
    ('system_event', 'api', '{"service": "auth", "status": "success"}', '{"service": "auth", "status": "success", "processed": true}', map('version', '1.0', 'region', 'us-west-2'), 8),
    ('error_event', 'batch', '{"error": "timeout"}', '{"error": "timeout", "severity": "warning", "processed": true}', map('version', '1.0', 'region', 'eu-west-1'), 23);

-- Create users and permissions
CREATE USER IF NOT EXISTS 'hoe_parser_user' IDENTIFIED BY 'hoe_parser_password';
GRANT SELECT, INSERT, CREATE, DROP ON hoe_parser.* TO 'hoe_parser_user';

-- Create read-only user for analytics
CREATE USER IF NOT EXISTS 'hoe_parser_readonly' IDENTIFIED BY 'readonly_password';
GRANT SELECT ON hoe_parser.* TO 'hoe_parser_readonly'; 