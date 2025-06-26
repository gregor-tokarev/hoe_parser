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
TTL created_at + INTERVAL 90 DAY;

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
TTL created_at + INTERVAL 30 DAY;

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
TTL created_at + INTERVAL 30 DAY;

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

-- ClickHouse initialization script for listings table
-- This script creates the main listings table with comprehensive pricing structure

-- Drop existing tables if they exist
DROP TABLE IF EXISTS listings;
DROP TABLE IF EXISTS listing_changes;

-- Main listings table with comprehensive pricing structure
CREATE TABLE listings (
    -- Primary identification
    id String,
    created_at DateTime64(3),
    updated_at DateTime64(3),
    last_scraped DateTime64(3),
    source_url String,
    
    -- Personal information
    personal_name String DEFAULT '',
    personal_age UInt8 DEFAULT 0,
    personal_height UInt16 DEFAULT 0,
    personal_weight UInt16 DEFAULT 0,
    personal_breast_size UInt8 DEFAULT 0,
    personal_hair_color String DEFAULT '',
    personal_eye_color String DEFAULT '',
    personal_body_type String DEFAULT '',
    
    -- Contact information
    contact_phone String DEFAULT '',
    contact_telegram String DEFAULT '',
    contact_email String DEFAULT '',
    
    -- Pricing currency
    pricing_currency String DEFAULT 'RUB',
    
    -- Structured pricing - Apartments/Incall Day rates
    price_apartments_day_hour UInt32 DEFAULT 0,
    price_apartments_day_2hour UInt32 DEFAULT 0,
    
    -- Structured pricing - Apartments/Incall Night rates  
    price_apartments_night_hour UInt32 DEFAULT 0,
    price_apartments_night_2hour UInt32 DEFAULT 0,
    
    -- Structured pricing - Outcall Day rates
    price_outcall_day_hour UInt32 DEFAULT 0,
    price_outcall_day_2hour UInt32 DEFAULT 0,
    
    -- Structured pricing - Outcall Night rates
    price_outcall_night_hour UInt32 DEFAULT 0,
    price_outcall_night_2hour UInt32 DEFAULT 0,
    
    -- Legacy/computed pricing fields for compatibility
    price_hour UInt32 DEFAULT 0,
    price_2_hours UInt32 DEFAULT 0,
    price_night UInt32 DEFAULT 0,
    price_day UInt32 DEFAULT 0,
    price_base UInt32 DEFAULT 0,
    
    -- Additional pricing data (for any other price types)
    pricing_duration_prices Map(String, UInt32) DEFAULT map(),
    pricing_service_prices Map(String, UInt32) DEFAULT map(),
    
    -- Service information
    service_available Array(String) DEFAULT [],
    service_additional Array(String) DEFAULT [],
    service_restrictions Array(String) DEFAULT [],
    service_meeting_type String DEFAULT '',
    
    -- Location information
    location_metro_stations Array(String) DEFAULT [],
    location_district String DEFAULT '',
    location_city String DEFAULT 'Unknown',
    location_outcall_available Bool DEFAULT false,
    location_incall_available Bool DEFAULT false,
    
    -- Content information
    description String DEFAULT '',
    last_updated String DEFAULT '',
    photos Array(String) DEFAULT [],
    photos_count UInt16 DEFAULT 0
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (id, location_city)
PARTITION BY toYYYYMM(created_at)
SETTINGS index_granularity = 8192;

-- Listing changes log table for tracking modifications
CREATE TABLE listing_changes (
    listing_id String,
    change_timestamp DateTime64(3) DEFAULT now64(),
    change_type String,
    field_name String,
    old_value String,
    new_value String,
    source String DEFAULT ''
) ENGINE = MergeTree()
ORDER BY (listing_id, change_timestamp)
PARTITION BY toYYYYMM(change_timestamp)
SETTINGS index_granularity = 8192;

-- Note: For querying latest listings, use "SELECT * FROM listings FINAL" in your queries

-- Indexes for better query performance
ALTER TABLE listings ADD INDEX idx_personal_age (personal_age) TYPE minmax GRANULARITY 1;
ALTER TABLE listings ADD INDEX idx_price_hour (price_hour) TYPE minmax GRANULARITY 1;
ALTER TABLE listings ADD INDEX idx_location_city (location_city) TYPE bloom_filter(0.01) GRANULARITY 1;
ALTER TABLE listings ADD INDEX idx_created_at (created_at) TYPE minmax GRANULARITY 1; 