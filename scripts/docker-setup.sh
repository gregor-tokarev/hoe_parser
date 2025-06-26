#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to display help
show_help() {
    echo -e "${GREEN}HOE Parser Docker Setup Script${NC}"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  up          Start all services"
    echo "  down        Stop all services"
    echo "  restart     Restart all services"
    echo "  logs        Show logs for all services"
    echo "  status      Show status of all services"
    echo "  clean       Clean up volumes and images"
    echo "  dev         Start in development mode"
    echo "  prod        Start in production mode"
    echo "  kafka       Open Kafka shell"
    echo "  clickhouse  Open ClickHouse client"
    echo "  build       Build the application image"
    echo "  test        Run tests in containers"
    echo "  help        Show this help message"
}

# Function to wait for service to be ready
wait_for_service() {
    local service=$1
    local port=$2
    local host=${3:-localhost}
    
    echo -e "${YELLOW}Waiting for $service to be ready...${NC}"
    
    while ! nc -z $host $port; do
        sleep 1
    done
    
    echo -e "${GREEN}$service is ready!${NC}"
}

# Function to create Kafka topics
create_kafka_topics() {
    echo -e "${YELLOW}Creating Kafka topics...${NC}"
    
    docker-compose exec kafka kafka-topics --create \
        --topic events \
        --partitions 3 \
        --replication-factor 1 \
        --if-not-exists \
        --bootstrap-server localhost:9092
    
    docker-compose exec kafka kafka-topics --create \
        --topic errors \
        --partitions 3 \
        --replication-factor 1 \
        --if-not-exists \
        --bootstrap-server localhost:9092
    
    docker-compose exec kafka kafka-topics --create \
        --topic metrics \
        --partitions 1 \
        --replication-factor 1 \
        --if-not-exists \
        --bootstrap-server localhost:9092
    
    echo -e "${GREEN}Kafka topics created successfully!${NC}"
}

# Function to start services
start_services() {
    local mode=${1:-prod}
    
    echo -e "${GREEN}Starting HOE Parser services in $mode mode...${NC}"
    
    if [ "$mode" = "dev" ]; then
        docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d
    else
        docker-compose up -d
    fi
    
    # Wait for services to be ready
    wait_for_service "Kafka" 9092
    wait_for_service "ClickHouse" 8123
    wait_for_service "Redis" 6379
    
    # Create Kafka topics
    create_kafka_topics
    
    echo -e "${GREEN}All services are up and running!${NC}"
    echo ""
    echo -e "${BLUE}Service URLs:${NC}"
    echo "  - HOE Parser API: http://localhost:8081"
    echo "  - Kafka UI: http://localhost:8080"
    echo "  - ClickHouse: http://localhost:8123"
    echo "  - Redis: localhost:6379"
}

# Function to stop services
stop_services() {
    echo -e "${YELLOW}Stopping HOE Parser services...${NC}"
    docker-compose down
    echo -e "${GREEN}Services stopped successfully!${NC}"
}

# Function to restart services
restart_services() {
    echo -e "${YELLOW}Restarting HOE Parser services...${NC}"
    stop_services
    start_services
}

# Function to show logs
show_logs() {
    local service=${1:-}
    
    if [ -n "$service" ]; then
        docker-compose logs -f "$service"
    else
        docker-compose logs -f
    fi
}

# Function to show status
show_status() {
    echo -e "${GREEN}Service Status:${NC}"
    docker-compose ps
}

# Function to clean up
cleanup() {
    echo -e "${YELLOW}Cleaning up Docker resources...${NC}"
    
    # Stop and remove containers
    docker-compose down -v --remove-orphans
    
    # Remove unused images
    docker image prune -f
    
    # Remove unused volumes
    docker volume prune -f
    
    echo -e "${GREEN}Cleanup completed!${NC}"
}

# Function to open Kafka shell
kafka_shell() {
    echo -e "${GREEN}Opening Kafka shell...${NC}"
    docker-compose exec kafka bash
}

# Function to open ClickHouse client
clickhouse_client() {
    echo -e "${GREEN}Opening ClickHouse client...${NC}"
    docker-compose exec clickhouse clickhouse-client --user admin --password password
}

# Function to build application
build_app() {
    echo -e "${GREEN}Building HOE Parser application...${NC}"
    docker-compose build hoe_parser
    echo -e "${GREEN}Build completed!${NC}"
}

# Function to run tests
run_tests() {
    echo -e "${GREEN}Running tests in containers...${NC}"
    
    # Start test environment
    docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d kafka clickhouse redis
    
    # Wait for services
    wait_for_service "Kafka" 9092
    wait_for_service "ClickHouse" 8123
    wait_for_service "Redis" 6379
    
    # Run tests
    docker-compose run --rm hoe_parser go test -v ./...
    
    echo -e "${GREEN}Tests completed!${NC}"
}

# Main script logic
case "${1:-}" in
    up)
        start_services prod
        ;;
    down)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    logs)
        show_logs "${2:-}"
        ;;
    status)
        show_status
        ;;
    clean)
        cleanup
        ;;
    dev)
        start_services dev
        ;;
    prod)
        start_services prod
        ;;
    kafka)
        kafka_shell
        ;;
    clickhouse)
        clickhouse_client
        ;;
    build)
        build_app
        ;;
    test)
        run_tests
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}Unknown command: ${1:-}${NC}"
        echo ""
        show_help
        exit 1
        ;;
esac 