#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building HOE Parser...${NC}"

# Get version from git tag or use default
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo -e "${YELLOW}Version: ${VERSION}${NC}"
echo -e "${YELLOW}Build Time: ${BUILD_TIME}${NC}"
echo -e "${YELLOW}Commit: ${COMMIT_HASH}${NC}"

# Build flags
LDFLAGS="-w -s"
LDFLAGS="${LDFLAGS} -X main.version=${VERSION}"
LDFLAGS="${LDFLAGS} -X main.buildTime=${BUILD_TIME}"
LDFLAGS="${LDFLAGS} -X main.commitHash=${COMMIT_HASH}"

# Create build directory
mkdir -p build

# Build for current platform
echo -e "${GREEN}Building for current platform...${NC}"
go build -ldflags "${LDFLAGS}" -o build/hoe_parser ./cmd/hoe_parser

# Build for multiple platforms
echo -e "${GREEN}Building for multiple platforms...${NC}"

platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    
    output_name="hoe_parser-${GOOS}-${GOARCH}"
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi
    
    echo -e "${YELLOW}Building for ${GOOS}/${GOARCH}...${NC}"
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "${LDFLAGS}" -o build/$output_name ./cmd/hoe_parser
done

echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}Binaries are available in the build/ directory${NC}" 