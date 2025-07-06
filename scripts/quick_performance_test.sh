#!/bin/bash

# GoNotes Quick Performance Test
# Simple load test for basic performance validation

set -e

echo "⚡ GoNotes Quick Performance Test"
echo "================================"

# Configuration
BASE_URL="http://localhost:8080"
CONCURRENT_USERS=20
TEST_DURATION=30
REQUESTS_PER_USER=50

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Function to print status
print_status() {
    local level=$1
    local message=$2
    
    case $level in
        "INFO") echo -e "${BLUE}[INFO]${NC} $message" ;;
        "SUCCESS") echo -e "${GREEN}[SUCCESS]${NC} $message" ;;
        "WARNING") echo -e "${YELLOW}[WARNING]${NC} $message" ;;
        "ERROR") echo -e "${RED}[ERROR]${NC} $message" ;;
    esac
}

# Check server availability
check_server() {
    print_status "INFO" "Checking server availability..."
    
    if curl -s -f "$BASE_URL/health" > /dev/null; then
        print_status "SUCCESS" "Server is running"
    else
        print_status "ERROR" "Server is not accessible at $BASE_URL"
        exit 1
    fi
}

# Setup test user
setup_test_user() {
    print_status "INFO" "Setting up test user..."
    
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local test_email="quicktest_$timestamp@example.com"
    local test_password="quicktest123"
    
    # Register user
    curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$test_email\",\"password\":\"$test_password\",\"full_name\":\"Quick Test User\"}" \
        "$BASE_URL/api/v1/auth/register" > /dev/null
    
    # Login and get token
    local login_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$test_email\",\"password\":\"$test_password\"}" \
        "$BASE_URL/api/v1/auth/login")
    
    ACCESS_TOKEN=$(echo "$login_response" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    
    if [[ -n "$ACCESS_TOKEN" ]]; then
        print_status "SUCCESS" "Test user ready"
        echo "Test email: $test_email"
    else
        print_status "ERROR" "Failed to setup test user"
        exit 1
    fi
}

# Run quick load test
run_quick_test() {
    local endpoint=$1
    local test_name=$2
    local use_auth=$3
    
    print_status "INFO" "Testing $test_name..."
    
    local start_time=$(date +%s)
    local successful=0
    local failed=0
    
    # Run concurrent requests
    for ((i=1; i<=CONCURRENT_USERS; i++)); do
        {
            for ((j=1; j<=10; j++)); do
                if [[ "$use_auth" == "true" ]]; then
                    response_code=$(curl -s -w "%{http_code}" -o /dev/null \
                        -H "Authorization: Bearer $ACCESS_TOKEN" \
                        "$BASE_URL$endpoint" 2>/dev/null)
                else
                    response_code=$(curl -s -w "%{http_code}" -o /dev/null \
                        "$BASE_URL$endpoint" 2>/dev/null)
                fi
                
                if [[ "$response_code" =~ ^[2][0-9][0-9]$ ]]; then
                    ((successful++))
                else
                    ((failed++))
                fi
                
                sleep 0.1
            done
        } &
        
        # Control concurrent processes
        if (( i % 5 == 0 )); then
            wait
        fi
    done
    
    wait  # Wait for all background jobs
    
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))
    local total_requests=$((successful + failed))
    local rps=0
    
    if [[ $total_time -gt 0 ]]; then
        rps=$(echo "scale=2; $total_requests / $total_time" | bc -l 2>/dev/null || echo "0")
    fi
    
    local success_rate=0
    if [[ $total_requests -gt 0 ]]; then
        success_rate=$(echo "scale=2; $successful * 100 / $total_requests" | bc -l 2>/dev/null || echo "0")
    fi
    
    # Display results
    echo "  📊 Results:"
    echo "     Total Requests: $total_requests"
    echo "     Successful: $successful"
    echo "     Failed: $failed"
    echo "     Success Rate: ${success_rate}%"
    echo "     Requests/Second: $rps"
    echo "     Test Duration: ${total_time}s"
    echo ""
}

# Main execution
main() {
    print_status "INFO" "Starting quick performance test..."
    
    # Check dependencies
    if ! command -v bc &> /dev/null; then
        print_status "WARNING" "bc not found. Some calculations may be inaccurate"
    fi
    
    # Initialize
    check_server
    setup_test_user
    
    # Run tests
    echo ""
    print_status "INFO" "Running load tests with $CONCURRENT_USERS concurrent users..."
    echo ""
    
    run_quick_test "/health" "Health Endpoint" "false"
    run_quick_test "/api/v1/user/profile" "Profile Endpoint" "true"
    run_quick_test "/api/v1/notes?page=1&page_size=10" "Notes Listing" "true"
    run_quick_test "/api/v1/notes/public?page=1&page_size=10" "Public Notes" "false"
    
    print_status "SUCCESS" "✅ Quick performance test completed!"
    
    # Performance recommendations
    echo ""
    print_status "INFO" "💡 Performance Tips:"
    echo "   - For detailed analysis, run: ./scripts/performance_test.sh"
    echo "   - Monitor Redis cache hit rates in production"
    echo "   - Consider connection pooling for high load"
    echo "   - Use CDN for static assets"
}

# Handle arguments
case "${1:-}" in
    "--help")
        echo "GoNotes Quick Performance Test"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help     Show this help message"
        echo ""
        echo "This script runs a quick load test with $CONCURRENT_USERS concurrent users"
        exit 0
        ;;
esac

# Execute main function
main "$@" 