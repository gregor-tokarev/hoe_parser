#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running tests for HOE Parser...${NC}"

# Run tests with coverage
echo -e "${YELLOW}Running unit tests...${NC}"
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
echo -e "${YELLOW}Generating coverage report...${NC}"
go tool cover -html=coverage.out -o coverage.html

# Display coverage summary
echo -e "${YELLOW}Coverage summary:${NC}"
go tool cover -func=coverage.out | tail -1

# Run benchmarks
echo -e "${YELLOW}Running benchmarks...${NC}"
go test -bench=. -benchmem ./...

# Vet the code
echo -e "${YELLOW}Running go vet...${NC}"
go vet ./...

# Check formatting
echo -e "${YELLOW}Checking code formatting...${NC}"
if [ "$(gofmt -l . | wc -l)" -gt 0 ]; then
    echo -e "${RED}Code is not properly formatted. Run 'go fmt ./...' to fix.${NC}"
    gofmt -l .
    exit 1
fi

echo -e "${GREEN}All tests passed!${NC}"
echo -e "${GREEN}Coverage report generated: coverage.html${NC}" 