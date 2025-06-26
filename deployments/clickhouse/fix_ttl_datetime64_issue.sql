-- Comprehensive fix for DateTime64 TTL issue
-- This script fixes all tables that have TTL expressions using DateTime64 columns

USE hoe_parser;

-- Drop materialized views first (they depend on tables)
DROP VIEW IF EXISTS listings_stats_mv;
DROP VIEW IF EXISTS events_summary_mv;

-- Drop existing tables (they will be recreated with correct TTL)
DROP TABLE IF EXISTS listing_changes;
DROP TABLE IF EXISTS listing_stats_daily;
DROP TABLE IF EXISTS listings;
DROP TABLE IF EXISTS parsing_errors;
DROP TABLE IF EXISTS metrics;
DROP TABLE IF EXISTS events;

-- Recreate events table with correct TTL
CREATE TABLE events (
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

-- Recreate metrics table with correct TTL
CREATE TABLE metrics (
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

-- Recreate parsing_errors table with correct TTL
CREATE TABLE parsing_errors (
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

-- Recreate listings table
CREATE TABLE listings (
    -- Primary identification
    id String,
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now(),
    last_scraped DateTime DEFAULT now(),
    source_url String,
    
    -- Personal information (flattened from PersonalInfo)
    personal_name String DEFAULT '',
    personal_age UInt8 DEFAULT 0,
    personal_height UInt16 DEFAULT 0,
    personal_weight UInt16 DEFAULT 0,
    personal_breast_size UInt8 DEFAULT 0,
    personal_hair_color String DEFAULT '',
    personal_eye_color String DEFAULT '',
    personal_body_type String DEFAULT '',
    
    -- Contact information (flattened from ContactInfo)
    contact_phone String DEFAULT '',
    contact_telegram String DEFAULT '',
    contact_email String DEFAULT '',
    contact_whatsapp_available Bool DEFAULT false,
    contact_viber_available Bool DEFAULT false,
    
    -- Pricing information (flattened from PricingInfo)
    pricing_currency String DEFAULT 'RUB',
    pricing_duration_prices Map(String, UInt32) DEFAULT map(),
    pricing_service_prices Map(String, UInt32) DEFAULT map(),
    
    -- Common pricing fields (extracted from maps for easier querying)
    price_hour UInt32 DEFAULT 0,
    price_2_hours UInt32 DEFAULT 0,
    price_night UInt32 DEFAULT 0,
    price_day UInt32 DEFAULT 0,
    price_base UInt32 DEFAULT 0,
    
    -- Service information (flattened from ServiceInfo)
    service_available Array(String) DEFAULT [],
    service_additional Array(String) DEFAULT [],
    service_restrictions Array(String) DEFAULT [],
    service_meeting_type String DEFAULT '',
    
    -- Location information (flattened from LocationInfo)
    location_metro_stations Array(String) DEFAULT [],
    location_district String DEFAULT '',
    location_city String DEFAULT '',
    location_outcall_available Bool DEFAULT false,
    location_incall_available Bool DEFAULT false,
    
    -- General information
    description String DEFAULT '',
    last_updated String DEFAULT '',
    photos Array(String) DEFAULT [],
    photos_count UInt16 DEFAULT 0,
    
    -- Search and analytics fields
    description_length UInt32 MATERIALIZED length(description),
    has_phone Bool MATERIALIZED length(contact_phone) > 0,
    has_telegram Bool MATERIALIZED length(contact_telegram) > 0,
    has_photos Bool MATERIALIZED length(photos) > 0,
    age_group String MATERIALIZED CASE 
        WHEN personal_age = 0 THEN 'unknown'
        WHEN personal_age < 20 THEN 'teen'
        WHEN personal_age < 25 THEN '20-24'
        WHEN personal_age < 30 THEN '25-29'
        WHEN personal_age < 35 THEN '30-34'
        WHEN personal_age < 40 THEN '35-39'
        ELSE '40+'
    END,
    price_range String MATERIALIZED CASE
        WHEN price_hour = 0 THEN 'unknown'
        WHEN price_hour < 5000 THEN 'budget'
        WHEN price_hour < 10000 THEN 'medium'
        WHEN price_hour < 20000 THEN 'premium'
        ELSE 'luxury'
    END
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (id, location_city)
PARTITION BY (toYYYYMM(created_at), location_city)
SETTINGS index_granularity = 8192;

-- Recreate listing_changes table
CREATE TABLE listing_changes (
    id UUID DEFAULT generateUUIDv4(),
    listing_id String,
    change_type String,
    old_value String DEFAULT '',
    new_value String DEFAULT '',
    field_name String DEFAULT '',
    timestamp DateTime DEFAULT now(),
    source String DEFAULT ''
) ENGINE = MergeTree()
ORDER BY (timestamp, listing_id, change_type)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 180 DAY;

-- Recreate listing_stats_daily table
CREATE TABLE listing_stats_daily (
    date Date,
    city String,
    total_listings UInt32,
    new_listings UInt32,
    updated_listings UInt32,
    avg_price_hour Float32,
    median_age Float32,
    created_at DateTime DEFAULT now()
) ENGINE = SummingMergeTree()
ORDER BY (date, city)
PARTITION BY toYYYYMM(date);

-- Recreate indexes for listings table
CREATE INDEX idx_age ON listings (personal_age) TYPE minmax GRANULARITY 1;
CREATE INDEX idx_price_hour ON listings (price_hour) TYPE minmax GRANULARITY 1;
CREATE INDEX idx_city ON listings (location_city) TYPE set(100) GRANULARITY 1;
CREATE INDEX idx_metro ON listings (location_metro_stations) TYPE bloom_filter(0.01) GRANULARITY 1;
CREATE INDEX idx_services ON listings (service_available) TYPE bloom_filter(0.01) GRANULARITY 1;

-- Recreate materialized views
CREATE MATERIALIZED VIEW events_summary_mv
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

CREATE MATERIALIZED VIEW listings_stats_mv
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

-- Grant permissions
GRANT SELECT, INSERT, CREATE, DROP ON hoe_parser.* TO 'hoe_parser_user';
GRANT SELECT, INSERT, UPDATE ON hoe_parser.listings TO 'hoe_parser_user';
GRANT SELECT, INSERT ON hoe_parser.listing_changes TO 'hoe_parser_user';
GRANT SELECT, INSERT ON hoe_parser.listing_stats_daily TO 'hoe_parser_user';

GRANT SELECT ON hoe_parser.* TO 'hoe_parser_readonly';
GRANT SELECT ON hoe_parser.listings TO 'hoe_parser_readonly';
GRANT SELECT ON hoe_parser.listing_changes TO 'hoe_parser_readonly';
GRANT SELECT ON hoe_parser.listing_stats_daily TO 'hoe_parser_readonly';

-- Verify the fix
SELECT 'All tables recreated with correct TTL expressions' as status; 