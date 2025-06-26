#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}HOE Parser Demo Script${NC}"
echo ""

# Build the applications
echo -e "${YELLOW}Building applications...${NC}"
make build
make scraper

echo ""
echo -e "${GREEN}Demo Options:${NC}"
echo "1. Run scraper example (direct scraping)"
echo "2. Start HTTP API server"
echo "3. Test HTTP API (requires server to be running)"
echo ""

read -p "Choose an option (1-3): " choice

case $choice in
    1)
        echo -e "${BLUE}Running scraper example...${NC}"
        echo "This will scrape the provided URL and display the results"
        echo ""
        ./build/scraper_example
        ;;
    2)
        echo -e "${BLUE}Starting HTTP API server...${NC}"
        echo "Server will start on http://localhost:8080"
        echo ""
        echo "Available endpoints:"
        echo "  GET  /health - Health check"
        echo "  POST /api/v1/scrape - Scrape a listing"
        echo ""
        echo "Press Ctrl+C to stop the server"
        echo ""
        ./build/hoe_parser
        ;;
    3)
        echo -e "${BLUE}Testing HTTP API...${NC}"
        echo "Testing health endpoint..."
        
        # Test health endpoint
        if curl -s http://localhost:8080/health > /dev/null; then
            echo -e "${GREEN}✓ Server is running${NC}"
            curl -s http://localhost:8080/health | jq .
            echo ""
            
            echo "Testing scrape endpoint..."
            curl -X POST http://localhost:8080/api/v1/scrape \
                -H "Content-Type: application/json" \
                -d '{"url": "https://a.intimcity.gold/indi/anketa675508.htm"}' \
                | jq .
        else
            echo -e "${RED}✗ Server is not running${NC}"
            echo "Please start the server first (option 2)"
        fi
        ;;
    *)
        echo -e "${RED}Invalid option${NC}"
        exit 1
        ;;
esac 