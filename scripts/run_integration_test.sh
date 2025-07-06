#!/bin/bash

# Integration Test Runner Script
# This script sets up the environment and runs integration tests

set -e

echo "🚀 GoNotes Integration Test Runner"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local level=$1
    local message=$2
    
    case $level in
        "INFO")
            echo -e "${BLUE}[INFO]${NC} $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} $message"
            ;;
        "WARNING")
            echo -e "${YELLOW}[WARNING]${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message"
            ;;
    esac
}

# Function to check if a service is running
check_service() {
    local service_name=$1
    local url=$2
    local max_attempts=30
    local attempt=0
    
    print_status "INFO" "Checking if $service_name is ready..."
    
    while [[ $attempt -lt $max_attempts ]]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            print_status "SUCCESS" "$service_name is ready"
            return 0
        fi
        attempt=$((attempt + 1))
        echo "Waiting for $service_name... ($attempt/$max_attempts)"
        sleep 2
    done
    
    print_status "ERROR" "$service_name failed to start"
    return 1
}

# Function to cleanup on exit
cleanup() {
    print_status "INFO" "Cleaning up..."
    
    # Kill Go application if running
    pkill -f "go run cmd/main.go" 2>/dev/null || true
    
    # Stop Docker containers
    docker-compose -f docker-compose.dev.yaml down 2>/dev/null || true
    
    print_status "SUCCESS" "Cleanup completed"
}

# Set trap for cleanup
trap cleanup EXIT

print_status "INFO" "Starting Docker services..."
# Start database and Redis
docker-compose -f docker-compose.dev.yaml up -d db redis

# Wait for services to be ready
sleep 5
print_status "SUCCESS" "Database and Redis are ready"

print_status "INFO" "Setting up environment..."
# Setup environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=gonotes_dev
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
export JWT_SECRET=dev_supersecretkey_for_testing_only
export JWT_EXPIRE=15m
export REFRESH_EXPIRE=7d
export APP_PORT=8080
export APP_ENV=development
export LOG_LEVEL=debug

print_status "INFO" "Starting GoNotes application..."
# Start the application in background
go run cmd/main.go &
APP_PID=$!

# Wait for application to start
check_service "GoNotes API" "http://localhost:8080/health"

print_status "INFO" "Running integration tests..."
# Run the integration tests
./scripts/test_integration_docker.sh

# Get the exit code
TEST_EXIT_CODE=$?

# Kill the application
kill $APP_PID 2>/dev/null || true

if [[ $TEST_EXIT_CODE -eq 0 ]]; then
    print_status "SUCCESS" "Integration tests completed successfully!"
else
    print_status "ERROR" "Integration tests failed!"
fi

exit $TEST_EXIT_CODE 