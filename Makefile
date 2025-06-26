.PHONY: build test clean lint fmt vet deps run docker-build docker-run docker-up docker-down docker-dev docker-status docker-logs docker-clean proto scraper help

# Variables
BINARY_NAME=hoe_parser
SCRAPER_BINARY=scraper_example
BUILD_DIR=build
CMD_DIR=cmd/hoe_parser
SCRAPER_CMD_DIR=cmd/scraper_example

# Default target
all: fmt vet test build

## Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

## Build the scraper example
scraper:
	@echo "Building $(SCRAPER_BINARY)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(SCRAPER_BINARY) ./$(SCRAPER_CMD_DIR)

## Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	@./scripts/generate-proto.sh

## Run tests
test:
	@echo "Running tests..."
	@./scripts/test.sh

## Build for multiple platforms
build-all:
	@echo "Building for all platforms..."
	@./scripts/build.sh

## Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

## Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## Vet code
vet:
	@echo "Vetting code..."
	@go vet ./...

## Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

## Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

## Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	@go run ./$(CMD_DIR)

## Run the scraper example
run-scraper:
	@echo "Running $(SCRAPER_BINARY)..."
	@go run ./$(SCRAPER_CMD_DIR)

## Run with hot reload (requires air)
dev:
	@echo "Running with hot reload..."
	@air -c .air.toml

## Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):latest .

## Docker run
docker-run:
	@echo "Running Docker container..."
	@docker run --rm -it -p 8080:8080 $(BINARY_NAME):latest

## Start all services with Docker Compose
docker-up:
	@echo "Starting all services..."
	@./scripts/docker-setup.sh up

## Stop all services
docker-down:
	@echo "Stopping all services..."
	@./scripts/docker-setup.sh down

## Start services in development mode
docker-dev:
	@echo "Starting services in development mode..."
	@./scripts/docker-setup.sh dev

## Show Docker services status
docker-status:
	@echo "Showing services status..."
	@./scripts/docker-setup.sh status

## Show Docker services logs
docker-logs:
	@echo "Showing services logs..."
	@./scripts/docker-setup.sh logs

## Clean Docker resources
docker-clean:
	@echo "Cleaning Docker resources..."
	@./scripts/docker-setup.sh clean

## Generate mocks (requires mockgen)
mocks:
	@echo "Generating mocks..."
	@go generate ./...

## Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golang/mock/mockgen@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## Show help
help:
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST) 