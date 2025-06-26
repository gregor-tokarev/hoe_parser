#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Generating Go code from protobuf files...${NC}"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${YELLOW}protoc is not installed. Please install it first.${NC}"
    echo "On macOS: brew install protobuf"
    echo "On Linux: apt-get install protobuf-compiler"
    exit 1
fi

# Create output directory
mkdir -p proto/listing

# Generate Go code from proto files
echo -e "${YELLOW}Generating listing.proto...${NC}"
protoc --go_out=. --go_opt=paths=source_relative proto/listing.proto

echo -e "${GREEN}Protobuf generation completed successfully!${NC}"
echo -e "${GREEN}Generated files:${NC}"
find proto -name "*.pb.go" -type f 