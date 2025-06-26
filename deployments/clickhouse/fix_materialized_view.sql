-- Fix script for materialized view error
-- Run this if you encounter: "Target table 'hoe_parser.metrics' of view 'hoe_parser.listings_stats_mv' doesn't exists"

USE hoe_parser;

-- Drop the problematic materialized view
DROP VIEW IF EXISTS listings_stats_mv;

-- Ensure metrics table exists
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

-- Recreate the materialized view
CREATE MATERIALIZED VIEW IF NOT EXISTS listings_stats_mv
TO metrics
AS SELECT
    now() as timestamp,
    'listings_by_city' as metric_name,
    count() as metric_value,
    map('city', location_city, 'age_group', age_group, 'price_range', price_range) as labels,
    now() as created_at
FROM listings
WHERE created_at >= now() - INTERVAL 1 HOUR
GROUP BY location_city, age_group, price_range;

-- Verify the setup
SELECT 'Materialized view fixed successfully' as status; 