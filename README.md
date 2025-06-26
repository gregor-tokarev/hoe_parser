# HOE Parser

A high-performance parsing service built with Go.

## 🚀 Features

- High-performance parsing engine
- RESTful API with OpenAPI documentation
- Docker containerization
- Comprehensive test coverage
- Multi-platform builds
- Configuration management
- Structured logging

## 📁 Project Structure

```
hoe_parser/
├── cmd/                    # Main applications
│   └── hoe_parser/        # Main application entry point
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   └── parser/           # Core parsing logic
├── pkg/                   # Public library code
│   └── utils/            # Utility functions
├── api/                   # API definitions
├── configs/               # Configuration files
├── scripts/               # Build and deployment scripts
├── build/                 # Build artifacts (generated)
├── docs/                  # Documentation
├── examples/              # Example code and usage
├── test/                  # Additional test files
└── deployments/           # Deployment configurations
```

## 🛠️ Development

### Prerequisites

- Go 1.24.3 or higher
- Make (optional, for using Makefile targets)
- Docker (optional, for containerization)

### Quick Start

1. **Clone the repository:**
   ```bash
   git clone https://github.com/gregor-tokarev/hoe_parser.git
   cd hoe_parser
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Build the application:**
   ```bash
   make build
   # or
   go build -o build/hoe_parser ./cmd/hoe_parser
   ```

4. **Run the application:**
   ```bash
   make run
   # or
   ./build/hoe_parser
   ```

### Available Make Targets

```bash
# Development
make help           # Show all available targets
make build          # Build the application
make scraper        # Build the scraper example
make proto          # Generate protobuf files
make test           # Run tests with coverage
make fmt            # Format code
make vet            # Vet code
make lint           # Lint code (requires golangci-lint)
make clean          # Clean build artifacts
make deps           # Install dependencies
make run            # Run the application
make run-scraper    # Run the scraper example
make dev            # Run with hot reload

# Docker
make docker-build   # Build Docker image
make docker-run     # Run single Docker container
make docker-up      # Start all services
make docker-down    # Stop all services
make docker-dev     # Start in development mode
make docker-status  # Show services status
make docker-logs    # Show services logs
make docker-clean   # Clean Docker resources
```

### Configuration

The application can be configured using:

1. **Environment variables:**
   ```bash
   export HOST=localhost
   export PORT=8080
   export LOG_LEVEL=info
   export DEBUG=false
   ```

2. **Configuration file:**
   Copy `configs/config.yaml` and modify as needed.

## 🧪 Testing

Run the test suite:

```bash
make test
```

This will:
- Run all unit tests with race detection
- Generate coverage report
- Run benchmarks
- Perform code vetting
- Check code formatting

## 🐳 Docker

### Build Docker image:
```bash
make docker-build
```

### Run with Docker:
```bash
make docker-run
```

### Using Docker Compose:
```bash
# Start all services (production mode)
make docker-up
# or
./scripts/docker-setup.sh up

# Start in development mode with hot reload
make docker-dev
# or
./scripts/docker-setup.sh dev

# Stop all services
make docker-down
# or
./scripts/docker-setup.sh down
```

## 🐳 Docker Services

The project includes a complete Docker Compose setup with the following services:

### Services Included:
- **HOE Parser Application** - Main application (port 8081)
- **Kafka** - Message broker (port 9092)
- **Kafka UI** - Web interface for Kafka management (port 8080)
- **ClickHouse** - Columnar database (HTTP: 8123, TCP: 9000)
- **Redis** - Caching and session storage (port 6379)
- **Zookeeper** - Required for Kafka (port 2181)

### Quick Commands:
```bash
# Start all services
make docker-up

# Start in development mode
make docker-dev

# Check status
make docker-status

# View logs
make docker-logs

# Stop all services
make docker-down

# Clean up resources
make docker-clean

# Access ClickHouse client
./scripts/docker-setup.sh clickhouse

# Access Kafka shell
./scripts/docker-setup.sh kafka
```

### Service URLs:
- **HOE Parser API**: http://localhost:8081
- **Kafka UI**: http://localhost:8080  
- **ClickHouse**: http://localhost:8123
- **Redis**: localhost:6379

## 🕷️ Web Scraping Functionality

The project includes advanced web scraping capabilities using [goquery](https://github.com/PuerkitoBio/goquery) and protobuf models:

### Features:
- **Structured Data Extraction**: Extracts personal info, contact details, pricing, services, and location
- **Protobuf Models**: Type-safe data structures generated from `.proto` definitions
- **HTTP API**: RESTful endpoints for scraping operations
- **Error Handling**: Robust error handling and validation

### Quick Start:
```bash
# Run the scraper example
make run-scraper

# Start the HTTP API server
make run

# Run the interactive demo
./scripts/demo.sh
```

### Protobuf Models:
The scraper returns structured data using these protobuf models:
- `Listing` - Main listing information
- `PersonalInfo` - Age, height, weight, physical attributes
- `ContactInfo` - Phone, telegram, messaging apps
- `PricingInfo` - Duration-based and service-based pricing
- `ServiceInfo` - Available services and restrictions
- `LocationInfo` - Metro stations, districts, availability

## 📚 API Documentation

The API is documented using OpenAPI 3.0. You can find the specification in `api/openapi.yaml`.

### Key Endpoints:

- `POST /api/v1/scrape` - Scrape and parse listing data
- `GET /api/v1/health` - Health check
- `GET /health` - Health check

### Example Usage:

```bash
# Health check
curl http://localhost:8080/health

# Scrape a listing
curl -X POST http://localhost:8080/api/v1/scrape \
  -H "Content-Type: application/json" \
  -d '{"url": "https://a.intimcity.gold/indi/anketa675508.htm"}'
```

## 🏗️ Architecture

### Directory Structure Explanation:

- **`cmd/`**: Contains the main applications. Each subdirectory represents a different binary.
- **`internal/`**: Private application code that shouldn't be imported by other applications.
- **`pkg/`**: Library code that can be used by external applications.
- **`api/`**: API definitions, protocol buffers, OpenAPI specs.
- **`configs/`**: Configuration file templates and examples.
- **`scripts/`**: Scripts for building, testing, and deployment.
- **`build/`**: Generated build artifacts.

### Package Organization:

- **Config Package**: Handles environment and file-based configuration
- **Parser Package**: Core business logic for parsing operations
- **Utils Package**: Reusable utility functions

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Run the test suite (`make test`)
6. Commit your changes (`git commit -m 'Add some amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Standards:

- Follow Go conventions and idioms
- Write tests for new functionality
- Ensure code is properly formatted (`make fmt`)
- Pass all linting checks (`make lint`)
- Maintain test coverage above 80%

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