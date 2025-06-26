# Makefile Environment Loading

All `run-*` commands in the Makefile automatically load environment variables from the `.env` file if it exists.

## How it Works

When you run any of the following commands, the Makefile will:
1. Check if a `.env` file exists in the project root
2. If found, load all environment variables from the file
3. Export them to the running process
4. Execute the Go command with the loaded environment

## Supported Commands

The following Makefile targets automatically load the `.env` file:

- `make run` - Run the main application
- `make run-scraper` - Run the basic scraper example
- `make run-gold-scraper` - Run the continuous gold scraper
- `make run-gold-callback-scraper` - Run the callback-based gold scraper
- `make run-clickhouse-example` - Run the ClickHouse integration example
- `make dev` - Run with hot reload using air

## Example Usage

### 1. Setup Environment File

```bash
# Copy the example environment file
cp env.example .env

# Edit your .env file with actual values
vim .env
```

### 2. Run with Environment

```bash
# The .env file will be automatically loaded
make run-clickhouse-example
```

You'll see output like:
```
Running clickhouse_example...
Loading environment from .env file...
Starting ClickHouse Adapter Example...
Loaded configuration: ClickHouse Host=localhost, Port=9000, Database=hoe_parser
✅ Connected to ClickHouse successfully!
```

### 3. Environment File Format

The `.env` file should contain key-value pairs:

```bash
# ClickHouse Configuration
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_DATABASE=hoe_parser
CLICKHOUSE_USER=hoe_parser_user
CLICKHOUSE_PASSWORD=hoe_parser_password

# Application Settings
DEBUG=true
LOG_LEVEL=debug
```

## Technical Implementation

Each run command uses this pattern:

```makefile
run-example:
	@echo "Running example..."
	@bash -c 'if [ -f .env ]; then echo "Loading environment from .env file..."; set -a; source .env; set +a; fi; go run ./cmd/example'
```

The command:
1. `bash -c` - Executes the entire command in a single bash shell
2. `if [ -f .env ]` - Checks if .env file exists
3. `set -a; source .env; set +a` - Loads and exports all variables from .env
4. `go run ./cmd/example` - Runs the Go application with loaded environment

## Benefits

- **Consistent Environment**: All run commands use the same environment setup
- **No Manual Export**: No need to manually export variables before running
- **Development Friendly**: Easy switching between different configurations
- **Docker Compatible**: Same .env file can be used with Docker Compose
- **Graceful Fallback**: Uses system environment if .env file doesn't exist

## Troubleshooting

### Environment Not Loading

If your environment variables aren't being loaded:

1. **Check .env file exists**:
   ```bash
   ls -la .env
   ```

2. **Check .env file format**:
   - No spaces around the `=` sign
   - No quotes needed unless value contains spaces
   - One variable per line

3. **Verify loading message**:
   You should see `Loading environment from .env file...` when running commands

### Example .env File Issues

❌ **Bad format**:
```bash
CLICKHOUSE_HOST = localhost  # spaces around =
CLICKHOUSE_PORT="9000"       # unnecessary quotes
```

✅ **Good format**:
```bash
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
```

## Security Notes

- **Never commit .env to version control** - It may contain sensitive credentials
- **Use env.example for documentation** - Commit this with example values
- **Set appropriate file permissions**: `chmod 600 .env` for sensitive files

## Integration with Configuration System

The environment variables loaded from `.env` are automatically used by the application's configuration system (`internal/config/config.go`), providing seamless integration between:

- Makefile environment loading
- Go application configuration
- Docker environment variables
- Production environment settings 