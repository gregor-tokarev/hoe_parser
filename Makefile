.PHONY: build test clean lint fmt vet deps run docker-build docker-run docker-up docker-down docker-dev docker-status docker-logs docker-clean proto help

# Variables
BINARY_NAME=hoe_parser
BUILD_DIR=build
CMD_DIR=cmd/hoe_parser

# Environment loading helper - loads .env file if it exists

# Default target
all: fmt vet test build

## Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

## Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	@./scripts/generate-proto.sh

## Run tests
test:
	@echo "Running tests..."
	@./scripts/test.sh

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