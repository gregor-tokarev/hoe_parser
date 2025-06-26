# HOE Parser

A high-performance parsing service built with Go.

## 🚀 Features

- High-performance parsing engine
- RESTful API with OpenAPI documentation
- **ClickHouse Integration** for analytics and data storage
- **Continuous Web Scraping** with real-time monitoring
- Docker containerization
- Comprehensive test coverage
- Multi-platform builds
- Configuration management
- Structured logging

## 📁 Project Structure

```
hoe_parser/
├── cmd/                    # Main applications
│   ├── hoe_parser/        # Main application entry point
│   ├── scraper_example/   # Basic scraper example
│   ├── intimcity_gold_example/     # Continuous gold scraper
│   ├── clickhouse_example/        # ClickHouse integration example
│   └── batch_to_clickhouse/       # Batch processing example
├── internal/              # Internal packages
│   ├── api/              # HTTP handlers and routes
│   ├── clickhouse/       # ClickHouse adapter and operations
│   ├── config/           # Configuration management
│   ├── kafka/            # Kafka client and operations
│   └── scraper/          # Web scraping functionality
├── deployments/          # Deployment configurations
│   └── clickhouse/       # ClickHouse setup and migrations
├── docs/                 # Documentation
│   ├── CLICKHOUSE_ADAPTER.md     # ClickHouse integration guide
│   └── INTIMCITY_GOLD_SCRAPER.md # Scraper documentation
├── pkg/                  # Public packages
├── proto/                # Protocol buffer definitions
└── scripts/              # Build and deployment scripts
```

## 🔧 Quick Start

### Prerequisites

- Go 1.24.3 or later
- Docker and Docker Compose
- ClickHouse (via Docker)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd hoe_parser

# Copy environment configuration
cp env.example .env

# Install dependencies
make deps

# Start services (ClickHouse, Kafka, etc.)
make docker-up

# Apply ClickHouse migrations (comprehensive setup)
docker exec -it clickhouse-server clickhouse-client --queries-file /docker-entrypoint-initdb.d/fix_ttl_datetime64_issue.sql

# Build the application
make build
```

### Examples

**Note**: All `make run-*` commands automatically load environment variables from `.env` file if it exists.

#### Basic Scraping
```bash
# Run basic scraper example
make run-scraper
```

#### Continuous Monitoring with ClickHouse
```bash
# Run continuous scraper with ClickHouse integration
make run-clickhouse-example
```

#### Batch Processing
```bash
# Run batch processing example
go run ./cmd/batch_to_clickhouse
```

## 📊 ClickHouse Integration

The project includes a comprehensive ClickHouse adapter for analytics and data storage:

- **Flattened Schema**: Optimized table structure for analytics
- **Batch Processing**: High-performance bulk operations  
- **Real-time Processing**: Continuous data ingestion with resilient error handling
- **Production-Ready Reliability**: Automatic retries, timeout management, and graceful error recovery
- **Built-in Analytics**: Statistics and reporting functions
- **Change Tracking**: Audit logging for all modifications

See [ClickHouse Adapter Documentation](docs/CLICKHOUSE_ADAPTER.md) for detailed information.

## 🔄 Web Scraping

Advanced web scraping capabilities with:

- **Continuous Monitoring**: Loop through pages automatically
- **Rate Limiting**: Respectful scraping with delays
- **Encoding Support**: Handle Russian text and various encodings
- **Error Recovery**: Robust error handling and retries
- **Channel Integration**: Real-time data processing via Go channels

See [Gold Scraper Documentation](docs/INTIMCITY_GOLD_SCRAPER.md) for detailed information.

## ⚙️ Configuration

The application uses environment variables for configuration. Key settings:

### ClickHouse
```bash
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_DATABASE=hoe_parser
CLICKHOUSE_USER=hoe_parser_user
CLICKHOUSE_PASSWORD=hoe_parser_password
```

### Application
```bash
HOST=localhost
PORT=8080
LOG_LEVEL=info
DEBUG=false
```

See `env.example` for all available configuration options.

## 🚀 Development

### Building

```bash
# Build all components
make build

# Build specific components
make clickhouse-example
make gold-scraper
```

### Testing

```bash
# Run tests
make test

# Run with coverage
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Vet code
make vet
```

## 📈 Monitoring

The application includes comprehensive monitoring:

- **Metrics**: Application and business metrics
- **Health Checks**: Service availability monitoring
- **Structured Logging**: JSON-formatted logs
- **Performance Profiling**: Built-in profiling support

## 🐳 Docker

### Development

```bash
# Start all services
make docker-dev

# Check service status
make docker-status

# View logs
make docker-logs
```

### Production

```bash
# Build Docker image
make docker-build

# Run container
make docker-run
```

## 📝 Available Commands

```bash
make help                    # Show all available commands
make build                   # Build the main application
make clickhouse-example      # Build ClickHouse example
make gold-scraper           # Build continuous scraper
make run-clickhouse-example # Run ClickHouse integration
make run-gold-scraper       # Run continuous scraper
make docker-up              # Start all services
make docker-down            # Stop all services
make test                   # Run tests
make clean                  # Clean build artifacts
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [Go Documentation](https://golang.org/doc/)
- [Project Layout Standards](https://github.com/golang-standards/project-layout)
- [OpenAPI Specification](https://spec.openapis.org/oas/v3.0.3)

## 📞 Support

If you have any questions or need help, please:

1. Check the documentation
2. Search existing issues
3. Create a new issue with a detailed description

---

Built with ❤️ using Go 