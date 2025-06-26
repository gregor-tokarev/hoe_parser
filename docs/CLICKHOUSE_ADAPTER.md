# ClickHouse Adapter

This adapter provides a comprehensive interface for storing and retrieving flattened listing data in ClickHouse. It converts the complex protobuf `Listing` structure into a flat table optimized for analytics and querying.

## Features

- **Automatic Flattening**: Converts nested protobuf structures to flat table columns
- **Batch Operations**: Efficient batch insertion for high-throughput scenarios
- **Change Tracking**: Logs all changes to listings for audit purposes
- **Statistics**: Built-in analytics and statistics functions
- **Type Safety**: Strongly-typed Go structs with proper ClickHouse mappings
- **Optimized Schema**: Designed for fast queries with proper indexes and partitioning
- **Configuration System**: Integrates with the main application configuration

## Setup

### 1. Environment Configuration

The adapter uses the main application configuration system. Configure ClickHouse connection via environment variables:

```bash
# Copy the example environment file
cp env.example .env

# Edit the ClickHouse configuration in .env
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_HTTP_PORT=8123
CLICKHOUSE_DATABASE=hoe_parser
CLICKHOUSE_USER=hoe_parser_user
CLICKHOUSE_PASSWORD=hoe_parser_password
CLICKHOUSE_MAX_CONNECTIONS=10
```

### 2. Run Migrations

Apply the ClickHouse migrations to create the necessary tables:

```bash
# Make sure ClickHouse is running via Docker Compose
make docker-up

# Method 1: Quick setup (recommended)
# Run the comprehensive fix script which recreates everything correctly
docker exec -it clickhouse-server clickhouse-client --queries-file /docker-entrypoint-initdb.d/fix_ttl_datetime64_issue.sql

# Method 2: Step-by-step setup
# Apply the base migration first (creates core tables)
docker exec -it clickhouse-server clickhouse-client --queries-file /docker-entrypoint-initdb.d/init.sql

# Apply the listings migration (creates listings table and views)
docker exec -it clickhouse-server clickhouse-client --queries-file /docker-entrypoint-initdb.d/listings_migration.sql
```

**Important**: 
- Use Method 1 for new setups or if you encounter TTL/materialized view errors
- Method 2 requires running `init.sql` before `listings_migration.sql` to ensure all dependencies are satisfied

### 3. Configuration in Code

The adapter integrates with the main configuration system:

```go
import (
    "github.com/gregor-tokarev/hoe_parser/internal/clickhouse"
    "github.com/gregor-tokarev/hoe_parser/internal/config"
)

// Load configuration from environment variables
cfg := config.Load()

// Create ClickHouse adapter configuration
chConfig := clickhouse.FromMainConfig(cfg, cfg.Debug)

// Create adapter
adapter, err := clickhouse.NewAdapter(chConfig)
if err != nil {
    log.Fatal(err)
}
defer adapter.Close()
```

## Usage Examples

### Basic Operations

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/gregor-tokarev/hoe_parser/internal/clickhouse"
    "github.com/gregor-tokarev/hoe_parser/internal/config"
    "github.com/gregor-tokarev/hoe_parser/internal/scraper"
)

func main() {
    // Load configuration from environment
    cfg := config.Load()
    chConfig := clickhouse.FromMainConfig(cfg, false)
    
    // Create adapter
    adapter, err := clickhouse.NewAdapter(chConfig)
    if err != nil {
        log.Fatal(err)
    }
    defer adapter.Close()

    // Single listing insertion
    scraper := scraper.NewIntimcityScraper()
    listing, err := scraper.ScrapeListing("https://example.com/listing123")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    err = adapter.InsertListing(ctx, listing, "https://example.com/listing123")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Batch Processing

```go
// Load configuration from environment
cfg := config.Load()
chConfig := clickhouse.FromMainConfig(cfg, false) // Set debug false for batch

// Create adapter
adapter, err := clickhouse.NewAdapter(chConfig)
if err != nil {
    log.Fatal(err)
}
defer adapter.Close()

// Batch insert multiple listings
var listings []*listing.Listing
var sourceURLs []string

// ... populate listings and sourceURLs ...

err := adapter.BatchInsertListings(ctx, listings, sourceURLs)
if err != nil {
    log.Fatal(err)
}
```

### Real-time Processing with Gold Scraper

```go
// Load configuration
cfg := config.Load()
chConfig := clickhouse.FromMainConfig(cfg, cfg.Debug)

// Create adapter
adapter, err := clickhouse.NewAdapter(chConfig)
if err != nil {
    log.Fatal(err)
}
defer adapter.Close()

// Continuous processing example
goldScraper := scraper.NewIntimcityGoldScraper()
linkChan := make(chan string, 100)

go func() {
    err := goldScraper.StartContinuousMonitoring(linkChan)
    if err != nil {
        log.Printf("Gold scraper error: %v", err)
    }
}()

intimcityScraper := scraper.NewIntimcityScraper()
for link := range linkChan {
    listing, err := intimcityScraper.ScrapeListing(link)
    if err != nil {
        continue
    }
    
    err = adapter.InsertListing(ctx, listing, link)
    if err != nil {
        log.Printf("Failed to insert: %v", err)
    }
}
```

## Error Handling and Resilience

The ClickHouse adapter includes robust error handling to ensure continuous operation even when individual operations fail:

### Retry Logic and Timeout Management

The updated examples include automatic retry logic with exponential backoff:

```go
// Function to retry ClickHouse operations
retryInsert := func(listing *listing.Listing, sourceURL string, maxRetries int) error {
    for attempt := 1; attempt <= maxRetries; attempt++ {
        // Create a context with timeout for this specific operation
        opCtx, opCancel := context.WithTimeout(ctx, 30*time.Second)
        
        err := adapter.InsertListing(opCtx, listing, sourceURL)
        opCancel()
        
        if err == nil {
            return nil
        }
        
        if attempt < maxRetries {
            log.Printf("⚠️  Attempt %d/%d failed for listing %s, retrying in %ds: %v", 
                attempt, maxRetries, listing.Id, attempt*2, err)
            time.Sleep(time.Duration(attempt*2) * time.Second)
        } else {
            return fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
        }
    }
    return nil
}
```

### Key Resilience Features

- **Isolated Timeouts**: Each operation has its own timeout context (30 seconds for inserts)
- **No Global Timeouts**: The main application runs indefinitely until manually stopped
- **Automatic Retries**: Failed insertions are retried up to 3 times with exponential backoff (2s, 4s, 6s delays)
- **Continue on Failure**: Individual failed operations don't stop the entire process
- **Detailed Monitoring**: Success/error rates are tracked and reported

### Monitoring Output

The real-time processing example provides comprehensive monitoring:

```
📥 Processing link #45: https://a.intimcity.gold/moskva/person123456
✅ Successfully inserted listing 123456 into ClickHouse (Success: 42/45)
📊 Stats: Processed=50, Success=47, Errors=3, Success Rate=94.0%
```

**Metrics Explained:**
- **Processed**: Total number of listings attempted
- **Success**: Successfully inserted listings (after retries)
- **Errors**: Operations that failed even after retries
- **Success Rate**: Percentage of successful operations

### Common Error Scenarios

#### Context Deadline Exceeded

**Symptoms:**
```
⚠️  Attempt 1/3 failed for listing 629234, retrying in 2s: failed to insert listing 629234: context deadline exceeded
✅ Successfully inserted listing 629234 into ClickHouse (Success: 43/50)
```

**Automatic Handling:**
- The adapter automatically retries with fresh timeout contexts
- Uses exponential backoff to reduce server load
- Continues processing other listings while retrying failed ones

**If Persistent Issues Occur:**
- Check ClickHouse server resources (CPU, memory, disk I/O)
- Verify network connectivity and latency
- Monitor ClickHouse query logs for slow queries
- Consider increasing timeout values or reducing batch sizes

#### Network Connectivity Issues

**Automatic Handling:**
- Each retry attempt creates a new connection context
- Failed connections don't affect subsequent operations
- Graceful degradation with detailed error logging

#### ClickHouse Server Overload

**Symptoms:**
```
⚠️  Attempt 1/3 failed for listing 123456, retrying in 2s: too many connections
⚠️  Attempt 2/3 failed for listing 123456, retrying in 4s: server busy
✅ Successfully inserted listing 123456 into ClickHouse (Success: 44/50)
```

**Automatic Handling:**
- Exponential backoff reduces load on overwhelmed servers
- Individual operation timeouts prevent resource starvation
- Connection pooling manages database connections efficiently

### Best Practices for Production

1. **Monitor Success Rates**: Watch for declining success rates that might indicate infrastructure issues
2. **Adjust Retry Limits**: Increase retry counts for unstable environments, decrease for stable ones
3. **Tune Timeouts**: Adjust operation timeouts based on your network and server performance
4. **Use Batch Processing**: For bulk operations, use `BatchInsertListings` which is more efficient
5. **Implement Alerting**: Set up alerts when error rates exceed acceptable thresholds

### Graceful Shutdown

The applications handle shutdown signals gracefully:

```go
// Setup signal handling
signalChan := make(chan os.Signal, 1)
signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

// ... processing code ...

// Wait for shutdown signal
<-signalChan
fmt.Println("\n🛑 Shutdown signal received. Stopping...")
cancel()

// Give goroutines time to clean up
time.Sleep(2 * time.Second)
fmt.Println("✅ Shutdown complete")
```

This ensures:
- Active operations can complete gracefully
- Database connections are properly closed
- No data loss during shutdown

## Configuration Reference

### Environment Variables

The adapter uses the following environment variables from the main configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `CLICKHOUSE_HOST` | `localhost` | ClickHouse server hostname |
| `CLICKHOUSE_PORT` | `9000` | ClickHouse native port |
| `CLICKHOUSE_HTTP_PORT` | `8123` | ClickHouse HTTP port |
| `CLICKHOUSE_DATABASE` | `hoe_parser` | Database name |
| `CLICKHOUSE_USER` | `admin` | Username for authentication |
| `CLICKHOUSE_PASSWORD` | `password` | Password for authentication |
| `CLICKHOUSE_MAX_CONNECTIONS` | `10` | Maximum connection pool size |
| `DEBUG` | `false` | Enable debug logging |

### Helper Functions

#### `FromMainConfig(mainCfg *config.Config, debug bool) Config`
Converts the main application configuration to a ClickHouse adapter configuration.

**Parameters:**
- `mainCfg`: Main application configuration
- `debug`: Enable debug mode for SQL logging

**Returns:** ClickHouse adapter configuration

## Build and Run Examples

### Environment Setup

```bash
# Set up environment variables
export CLICKHOUSE_HOST=localhost
export CLICKHOUSE_PORT=9000
export CLICKHOUSE_DATABASE=hoe_parser
export CLICKHOUSE_USER=hoe_parser_user
export CLICKHOUSE_PASSWORD=hoe_parser_password
export DEBUG=true
```

### Build and Run

```bash
# Build the ClickHouse adapter example
make clickhouse-example

# Run the continuous processing example
make run-clickhouse-example

# Build and run batch processing example
go build -o build/batch_to_clickhouse ./cmd/batch_to_clickhouse
./build/batch_to_clickhouse
```

## Docker Integration

The configuration system works seamlessly with Docker environments:

```yaml
# docker-compose.yml
services:
  hoe_parser:
    environment:
      - CLICKHOUSE_HOST=clickhouse-server
      - CLICKHOUSE_PORT=9000
      - CLICKHOUSE_DATABASE=hoe_parser
      - CLICKHOUSE_USER=hoe_parser_user
      - CLICKHOUSE_PASSWORD=hoe_parser_password
      - DEBUG=false
```

## Configuration Validation

The adapter automatically validates configuration on startup:

- Tests connection to ClickHouse server
- Verifies database accessibility
- Validates authentication credentials
- Reports configuration errors with helpful messages

Example output:
```
Loaded configuration: ClickHouse Host=localhost, Port=9000, Database=hoe_parser
✅ Connected to ClickHouse successfully!
```

## Table Schema

### Main Listings Table

The `listings` table contains flattened data with the following structure:

```sql
-- Primary identification
id String
created_at DateTime
updated_at DateTime
last_scraped DateTime
source_url String

-- Personal information (flattened from PersonalInfo)
personal_name String
personal_age UInt8
personal_height UInt16
personal_weight UInt16
personal_breast_size UInt8
personal_hair_color String
personal_eye_color String
personal_body_type String

-- Contact information (flattened from ContactInfo)
contact_phone String
contact_telegram String
contact_email String
contact_whatsapp_available Bool
contact_viber_available Bool

-- Pricing information (flattened from PricingInfo)
pricing_currency String
pricing_duration_prices Map(String, UInt32)
pricing_service_prices Map(String, UInt32)
price_hour UInt32
price_2_hours UInt32
price_night UInt32
price_day UInt32
price_base UInt32

-- Service information (flattened from ServiceInfo)
service_available Array(String)
service_additional Array(String)
service_restrictions Array(String)
service_meeting_type String

-- Location information (flattened from LocationInfo)
location_metro_stations Array(String)
location_district String
location_city String
location_outcall_available Bool
location_incall_available Bool

-- General information
description String
last_updated String
photos Array(String)
photos_count UInt16

-- Computed fields (MATERIALIZED)
description_length UInt32
has_phone Bool
has_telegram Bool
has_photos Bool
age_group String
price_range String
```

### Supporting Tables

- **`listing_changes`**: Audit log for all listing modifications
- **`listing_stats_daily`**: Daily aggregated statistics by city
- **`metrics`**: General metrics table (inherited from existing schema)

## API Reference

### Adapter Methods

#### `NewAdapter(config Config) (*Adapter, error)`
Creates a new ClickHouse adapter with the given configuration.

#### `Close() error`
Closes the ClickHouse connection.

#### `InsertListing(ctx context.Context, listing *listing.Listing, sourceURL string) error`
Inserts a single listing into ClickHouse.

#### `BatchInsertListings(ctx context.Context, listings []*listing.Listing, sourceURLs []string) error`
Batch inserts multiple listings for better performance.

#### `UpdateListing(ctx context.Context, listing *listing.Listing, sourceURL string) error`
Updates an existing listing (uses ClickHouse's ReplacingMergeTree for upsert behavior).

#### `GetListingByID(ctx context.Context, id string) (*FlattenedListing, error)`
Retrieves a listing by its ID.

#### `GetStats(ctx context.Context) (map[string]interface{}, error)`
Returns comprehensive statistics about the listings in the database.

#### `LogChange(ctx context.Context, listingID, changeType, oldValue, newValue, fieldName, source string) error`
Logs a change to the `listing_changes` table for audit purposes.

### Data Types

#### `FlattenedListing`
Go struct representing the flattened listing structure, with all nested protobuf fields expanded into flat fields.

#### `Config`
Configuration struct for ClickHouse connection parameters, compatible with the main application configuration.

### Configuration Helper

#### `FromMainConfig(mainCfg *config.Config, debug bool) Config`
Helper function to create adapter configuration from main application configuration.

## Performance Optimizations

### Table Engine
- **ReplacingMergeTree**: Automatically handles updates based on the `updated_at` field
- **Partitioning**: By month and city for efficient querying
- **Ordering**: By `(id, location_city)` for optimal performance

### Indexes
- **MinMax indexes**: On age and price fields for range queries
- **Set indexes**: On city for equality queries  
- **Bloom filter indexes**: On metro stations and services for array searches

### Batch Processing
- Use `BatchInsertListings()` for inserting multiple records
- Recommended batch size: 100-1000 records
- Automatic batching for high-throughput scenarios

## Query Examples

### Basic Analytics

```sql
-- Average price by city
SELECT 
    location_city,
    avg(price_hour) as avg_price,
    count() as total_listings
FROM listings 
WHERE price_hour > 0
GROUP BY location_city
ORDER BY avg_price DESC;

-- Age distribution
SELECT 
    age_group,
    count() as count,
    avg(price_hour) as avg_price
FROM listings
WHERE personal_age > 0
GROUP BY age_group
ORDER BY count DESC;

-- Price ranges
SELECT 
    price_range,
    count() as count,
    min(price_hour) as min_price,
    max(price_hour) as max_price
FROM listings
WHERE price_hour > 0
GROUP BY price_range;
```

### Advanced Queries

```sql
-- Listings with specific services
SELECT id, personal_age, price_hour, location_city
FROM listings
WHERE has(service_available, 'массаж')
AND location_city = 'Moscow'
AND price_hour BETWEEN 5000 AND 15000;

-- Metro station popularity
SELECT 
    metro_station,
    count() as listings_count,
    avg(price_hour) as avg_price
FROM listings
ARRAY JOIN location_metro_stations AS metro_station
WHERE metro_station != ''
GROUP BY metro_station
ORDER BY listings_count DESC
LIMIT 20;

-- Time-based analysis
SELECT 
    toYYYYMM(created_at) as month,
    count() as new_listings,
    avg(price_hour) as avg_price
FROM listings
GROUP BY month
ORDER BY month;
```

## Monitoring and Maintenance

### Materialized Views
The schema includes materialized views that automatically aggregate statistics into the `metrics` table.

### TTL Policies
- `listing_changes`: 180 days retention
- `metrics`: 30 days retention

### Manual Cleanup
```sql
-- Force merge to apply ReplacingMergeTree deduplication
OPTIMIZE TABLE listings FINAL;

-- Check table sizes
SELECT 
    table,
    formatReadableSize(sum(bytes)) as size
FROM system.parts
WHERE database = 'hoe_parser'
GROUP BY table;
```

## Error Handling

The adapter includes comprehensive error handling:
- Connection failures with retry logic
- Batch operation failures with partial recovery
- Type conversion errors with detailed logging
- Context timeout handling for long-running operations
- Configuration validation with helpful error messages

## Troubleshooting

### Materialized View Error

**Error**: `Target table 'hoe_parser.metrics' of view 'hoe_parser.listings_stats_mv' doesn't exists`

**Cause**: The listings migration was applied before the base migration, or the `metrics` table was not created.

**Fix**:
```bash
# Run the fix script
docker exec -it clickhouse-server clickhouse-client --queries-file /docker-entrypoint-initdb.d/fix_materialized_view.sql

# Or manually fix:
docker exec -it clickhouse-server clickhouse-client --query "
USE hoe_parser;
DROP VIEW IF EXISTS listings_stats_mv;
"

# Then re-apply the corrected listings migration
docker exec -it clickhouse-server clickhouse-client --queries-file /docker-entrypoint-initdb.d/listings_migration.sql
```

### DateTime64 TTL Error

**Error**: `TTL expression result column should have DateTime or Date type, but has DateTime64(3). (BAD_TTL_EXPRESSION)`

**Cause**: ClickHouse TTL expressions don't support DateTime64 columns, only DateTime or Date columns.

**Fix** (Complete database reset):
```bash
# Run the comprehensive fix script
docker exec -it clickhouse-server clickhouse-client --queries-file /docker-entrypoint-initdb.d/fix_ttl_datetime64_issue.sql
```

**Fix** (Manual correction):
```bash
# All TTL expressions have been updated to use DateTime columns instead of DateTime64
# The tables use both DateTime64(3) for high-precision timestamps and DateTime for TTL
docker exec -it clickhouse-server clickhouse-client --query "
USE hoe_parser;
-- Example: Fix metrics table TTL
ALTER TABLE metrics MODIFY TTL created_at + INTERVAL 30 DAY;
"
```

### Migration Order Issues

Always apply migrations in this order:
1. `init.sql` - Creates base tables (`metrics`, `events`, `parsing_errors`)
2. `listings_migration.sql` - Creates listings table and related structures

### Connection Issues

**Error**: `failed to ping ClickHouse: dial tcp: connect: connection refused`

**Fix**:
```bash
# Ensure ClickHouse is running
make docker-up

# Check ClickHouse status
docker ps | grep clickhouse

# Check ClickHouse logs
make docker-logs
```

### Permission Issues

**Error**: `Authentication failed`

**Fix**: Ensure your `.env` file has the correct credentials:
```bash
CLICKHOUSE_USER=hoe_parser_user
CLICKHOUSE_PASSWORD=hoe_parser_password
```

## Integration with Existing Systems

The ClickHouse adapter integrates seamlessly with:
- **Gold Scraper**: Real-time processing of new listings
- **Intimcity Scraper**: Individual listing processing
- **Main Configuration System**: Uses environment variables and config package
- **Kafka**: Can be extended to consume from message queues
- **API Layer**: Direct queries through existing API handlers
- **Docker/Container Environments**: Full support for containerized deployments 