-- Migration for listings table
-- This creates a flattened table structure from the Listing protobuf

USE hoe_parser;

-- Ensure metrics table exists (required for materialized views)
CREATE TABLE IF NOT EXISTS metrics (
    id UUID DEFAULT generateUUIDv4(),
    timestamp DateTime DEFAULT now(),
    metric_name String,
    metric_value Float64,
    labels Map(String, String),
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (timestamp, metric_name)
PARTITION BY toYYYYMM(timestamp)
TTL created_at + INTERVAL 30 DAY;

-- Create listings table with flattened structure
CREATE TABLE IF NOT EXISTS listings (
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
    pricing_duration_prices Map(String, UInt32) DEFAULT map(), -- duration -> price
    pricing_service_prices Map(String, UInt32) DEFAULT map(),  -- service -> additional price
    
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
    service_meeting_type String DEFAULT '', -- apartment, hotel, etc.
    
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

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_age ON listings (personal_age) TYPE minmax GRANULARITY 1;
CREATE INDEX IF NOT EXISTS idx_price_hour ON listings (price_hour) TYPE minmax GRANULARITY 1;
CREATE INDEX IF NOT EXISTS idx_city ON listings (location_city) TYPE set(100) GRANULARITY 1;
CREATE INDEX IF NOT EXISTS idx_metro ON listings (location_metro_stations) TYPE bloom_filter(0.01) GRANULARITY 1;
CREATE INDEX IF NOT EXISTS idx_services ON listings (service_available) TYPE bloom_filter(0.01) GRANULARITY 1;

-- Create materialized view for analytics
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

-- Create table for tracking listing updates/changes
CREATE TABLE IF NOT EXISTS listing_changes (
    id UUID DEFAULT generateUUIDv4(),
    listing_id String,
    change_type String, -- 'created', 'updated', 'price_changed', etc.
    old_value String DEFAULT '',
    new_value String DEFAULT '',
    field_name String DEFAULT '',
    timestamp DateTime DEFAULT now(),
    source String DEFAULT ''
) ENGINE = MergeTree()
ORDER BY (timestamp, listing_id, change_type)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 180 DAY;

-- Create aggregated statistics table
CREATE TABLE IF NOT EXISTS listing_stats_daily (
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

-- Insert initial test data (optional - remove in production)
INSERT INTO listings (
    id, source_url, personal_name, personal_age, personal_height, personal_weight, 
    personal_breast_size, contact_phone, pricing_currency, price_hour, 
    location_city, description, photos
) VALUES (
    'test_001', 
    'https://example.com/test1', 
    'Test User', 
    25, 
    165, 
    55, 
    3, 
    '+7900123456', 
    'RUB', 
    8000, 
    'Moscow', 
    'Test description for listing',
    ['photo1.jpg', 'photo2.jpg']
);

-- Grant permissions to existing users
GRANT SELECT, INSERT, UPDATE ON hoe_parser.listings TO 'hoe_parser_user';
GRANT SELECT, INSERT ON hoe_parser.listing_changes TO 'hoe_parser_user';
GRANT SELECT, INSERT ON hoe_parser.listing_stats_daily TO 'hoe_parser_user';

GRANT SELECT ON hoe_parser.listings TO 'hoe_parser_readonly';
GRANT SELECT ON hoe_parser.listing_changes TO 'hoe_parser_readonly';
GRANT SELECT ON hoe_parser.listing_stats_daily TO 'hoe_parser_readonly'; 