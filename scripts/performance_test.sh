#!/bin/bash

# GoNotes Performance Testing Script
# Comprehensive performance, load, and stress testing for GoNotes API

set -e

echo "🚀 GoNotes Performance Testing Suite"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080"
TEST_RESULTS_DIR="performance_results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="$TEST_RESULTS_DIR/performance_report_$TIMESTAMP.txt"

# Test configuration
LIGHT_LOAD_USERS=10
MEDIUM_LOAD_USERS=50
HEAVY_LOAD_USERS=100
STRESS_TEST_USERS=200
SPIKE_TEST_USERS=500

TEST_DURATION=60  # seconds
WARM_UP_TIME=10   # seconds
COOL_DOWN_TIME=5  # seconds

# Global variables
ACCESS_TOKEN=""
REFRESH_TOKEN=""
TEST_USER_COUNT=0

# Create results directory
mkdir -p "$TEST_RESULTS_DIR"

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
        "PERF")
            echo -e "${PURPLE}[PERF]${NC} $message"
            ;;
        "TEST")
            echo -e "${CYAN}[TEST]${NC} $message"
            ;;
    esac
}

# Function to log to report file
log_to_report() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$REPORT_FILE"
}

# Function to initialize report
init_report() {
    echo "GoNotes Performance Testing Report" > "$REPORT_FILE"
    echo "=================================" >> "$REPORT_FILE"
    echo "Timestamp: $(date)" >> "$REPORT_FILE"
    echo "Base URL: $BASE_URL" >> "$REPORT_FILE"
    echo "Test Configuration:" >> "$REPORT_FILE"
    echo "- Light Load: $LIGHT_LOAD_USERS concurrent users" >> "$REPORT_FILE"
    echo "- Medium Load: $MEDIUM_LOAD_USERS concurrent users" >> "$REPORT_FILE"
    echo "- Heavy Load: $HEAVY_LOAD_USERS concurrent users" >> "$REPORT_FILE"
    echo "- Stress Test: $STRESS_TEST_USERS concurrent users" >> "$REPORT_FILE"
    echo "- Spike Test: $SPIKE_TEST_USERS concurrent users" >> "$REPORT_FILE"
    echo "- Test Duration: $TEST_DURATION seconds" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
}

# Function to check if server is accessible
check_server() {
    print_status "INFO" "Checking server accessibility..."
    
    if curl -s -f "$BASE_URL/health" > /dev/null; then
        print_status "SUCCESS" "Server is accessible"
        return 0
    else
        print_status "ERROR" "Server is not accessible at $BASE_URL"
        exit 1
    fi
}

# Function to setup test user
setup_test_user() {
    print_status "INFO" "Setting up test user..."
    
    local test_email="perftest_$TIMESTAMP@example.com"
    local test_password="perftest123"
    local test_name="Performance Test User"
    
    # Register test user
    local register_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$test_email\",
            \"password\": \"$test_password\",
            \"full_name\": \"$test_name\"
        }" \
        "$BASE_URL/api/v1/auth/register" 2>/dev/null)
    
    # Login to get tokens
    local login_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$test_email\",
            \"password\": \"$test_password\"
        }" \
        "$BASE_URL/api/v1/auth/login" 2>/dev/null)
    
    # Extract tokens
    ACCESS_TOKEN=$(echo "$login_response" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo "$login_response" | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)
    
    if [[ -n "$ACCESS_TOKEN" && -n "$REFRESH_TOKEN" ]]; then
        print_status "SUCCESS" "Test user setup completed"
        log_to_report "Test user created: $test_email"
        return 0
    else
        print_status "ERROR" "Failed to setup test user"
        return 1
    fi
}

# Function to make concurrent HTTP requests
make_concurrent_requests() {
    local endpoint=$1
    local method=$2
    local concurrent_users=$3
    local duration=$4
    local request_data=$5
    local use_auth=$6
    
    local temp_results_file="$TEST_RESULTS_DIR/temp_results_$RANDOM.txt"
    local successful_requests=0
    local failed_requests=0
    local total_time=0
    local min_time=999999
    local max_time=0
    
    print_status "TEST" "Running $concurrent_users concurrent users for ${duration}s on $method $endpoint"
    
    # Warm up
    print_status "INFO" "Warming up for ${WARM_UP_TIME}s..."
    for ((i=1; i<=5; i++)); do
        curl -s "$BASE_URL$endpoint" > /dev/null 2>&1 || true
    done
    sleep $WARM_UP_TIME
    
    # Start concurrent requests
    local end_time=$((SECONDS + duration))
    local pids=()
    
    for ((i=1; i<=concurrent_users; i++)); do
        {
            while [[ $SECONDS -lt $end_time ]]; do
                local start_time=$(date +%s.%3N)
                
                if [[ "$use_auth" == "true" ]]; then
                    if [[ -n "$request_data" ]]; then
                        local response_code=$(curl -s -w "%{http_code}" -o /dev/null \
                            -X "$method" \
                            -H "Authorization: Bearer $ACCESS_TOKEN" \
                            -H "Content-Type: application/json" \
                            -d "$request_data" \
                            "$BASE_URL$endpoint" 2>/dev/null)
                    else
                        local response_code=$(curl -s -w "%{http_code}" -o /dev/null \
                            -X "$method" \
                            -H "Authorization: Bearer $ACCESS_TOKEN" \
                            "$BASE_URL$endpoint" 2>/dev/null)
                    fi
                else
                    if [[ -n "$request_data" ]]; then
                        local response_code=$(curl -s -w "%{http_code}" -o /dev/null \
                            -X "$method" \
                            -H "Content-Type: application/json" \
                            -d "$request_data" \
                            "$BASE_URL$endpoint" 2>/dev/null)
                    else
                        local response_code=$(curl -s -w "%{http_code}" -o /dev/null \
                            -X "$method" \
                            "$BASE_URL$endpoint" 2>/dev/null)
                    fi
                fi
                
                local end_request_time=$(date +%s.%3N)
                local request_time=$(echo "$end_request_time - $start_time" | bc -l 2>/dev/null || echo "0")
                
                echo "$response_code,$request_time" >> "$temp_results_file"
                
                # Small delay to prevent overwhelming
                sleep 0.01
            done
        } &
        pids+=($!)
    done
    
    # Wait for all background processes
    for pid in "${pids[@]}"; do
        wait "$pid"
    done
    
    # Cool down
    sleep $COOL_DOWN_TIME
    
    # Analyze results
    if [[ -f "$temp_results_file" ]]; then
        while IFS=',' read -r code time; do
            if [[ "$code" =~ ^[2][0-9][0-9]$ ]]; then
                ((successful_requests++))
            else
                ((failed_requests++))
            fi
            
            # Calculate timing statistics
            if [[ -n "$time" ]] && [[ "$time" != "0" ]]; then
                total_time=$(echo "$total_time + $time" | bc -l 2>/dev/null || echo "$total_time")
                if (( $(echo "$time < $min_time" | bc -l 2>/dev/null || echo "0") )); then
                    min_time=$time
                fi
                if (( $(echo "$time > $max_time" | bc -l 2>/dev/null || echo "0") )); then
                    max_time=$time
                fi
            fi
        done < "$temp_results_file"
        
        rm -f "$temp_results_file"
    fi
    
    local total_requests=$((successful_requests + failed_requests))
    local success_rate=0
    local avg_response_time=0
    local rps=0
    
    if [[ $total_requests -gt 0 ]]; then
        success_rate=$(echo "scale=2; $successful_requests * 100 / $total_requests" | bc -l 2>/dev/null || echo "0")
        rps=$(echo "scale=2; $total_requests / $duration" | bc -l 2>/dev/null || echo "0")
    fi
    
    if [[ $successful_requests -gt 0 ]] && [[ "$total_time" != "0" ]]; then
        avg_response_time=$(echo "scale=3; $total_time / $successful_requests" | bc -l 2>/dev/null || echo "0")
    fi
    
    # Display results
    print_status "PERF" "Results for $concurrent_users users on $method $endpoint:"
    echo "  Total Requests: $total_requests"
    echo "  Successful: $successful_requests"
    echo "  Failed: $failed_requests"
    echo "  Success Rate: ${success_rate}%"
    echo "  Requests/Second: $rps"
    echo "  Avg Response Time: ${avg_response_time}s"
    echo "  Min Response Time: ${min_time}s"
    echo "  Max Response Time: ${max_time}s"
    echo ""
    
    # Log to report
    log_to_report "Test: $concurrent_users users - $method $endpoint"
    log_to_report "  Total Requests: $total_requests"
    log_to_report "  Successful: $successful_requests"
    log_to_report "  Failed: $failed_requests"
    log_to_report "  Success Rate: ${success_rate}%"
    log_to_report "  Requests/Second: $rps"
    log_to_report "  Avg Response Time: ${avg_response_time}s"
    log_to_report "  Min Response Time: ${min_time}s"
    log_to_report "  Max Response Time: ${max_time}s"
    log_to_report ""
}

# Function to run health check performance test
test_health_performance() {
    print_status "TEST" "Testing Health Endpoint Performance"
    log_to_report "=== Health Endpoint Performance ==="
    
    make_concurrent_requests "/health" "GET" $LIGHT_LOAD_USERS $TEST_DURATION "" "false"
    make_concurrent_requests "/health" "GET" $MEDIUM_LOAD_USERS $TEST_DURATION "" "false"
    make_concurrent_requests "/health" "GET" $HEAVY_LOAD_USERS $TEST_DURATION "" "false"
}

# Function to run authentication performance test
test_auth_performance() {
    print_status "TEST" "Testing Authentication Performance"
    log_to_report "=== Authentication Performance ==="
    
    local login_data='{"email":"perftest_'$TIMESTAMP'@example.com","password":"perftest123"}'
    
    make_concurrent_requests "/api/v1/auth/login" "POST" $LIGHT_LOAD_USERS 30 "$login_data" "false"
    make_concurrent_requests "/api/v1/auth/login" "POST" $MEDIUM_LOAD_USERS 30 "$login_data" "false"
    
    # Test token refresh
    local refresh_data='{"refresh_token":"'$REFRESH_TOKEN'"}'
    make_concurrent_requests "/api/v1/auth/refresh" "POST" $LIGHT_LOAD_USERS 30 "$refresh_data" "false"
}

# Function to run profile performance test
test_profile_performance() {
    print_status "TEST" "Testing Profile Management Performance"
    log_to_report "=== Profile Management Performance ==="
    
    # Test profile retrieval (with Redis caching)
    make_concurrent_requests "/api/v1/user/profile" "GET" $LIGHT_LOAD_USERS $TEST_DURATION "" "true"
    make_concurrent_requests "/api/v1/user/profile" "GET" $MEDIUM_LOAD_USERS $TEST_DURATION "" "true"
    make_concurrent_requests "/api/v1/user/profile" "GET" $HEAVY_LOAD_USERS $TEST_DURATION "" "true"
    
    # Test profile updates
    local update_data='{"email":"perftest_'$TIMESTAMP'@example.com","full_name":"Updated Performance Test User"}'
    make_concurrent_requests "/api/v1/user/profile" "PUT" $LIGHT_LOAD_USERS 30 "$update_data" "true"
}

# Function to run notes performance test
test_notes_performance() {
    print_status "TEST" "Testing Notes System Performance"
    log_to_report "=== Notes System Performance ==="
    
    # Test notes listing
    make_concurrent_requests "/api/v1/notes?page=1&page_size=10" "GET" $LIGHT_LOAD_USERS $TEST_DURATION "" "true"
    make_concurrent_requests "/api/v1/notes?page=1&page_size=10" "GET" $MEDIUM_LOAD_USERS $TEST_DURATION "" "true"
    
    # Test note creation
    local note_data='{"title":"Performance Test Note","content":"This is a test note for performance testing","tags":["performance","test"]}'
    make_concurrent_requests "/api/v1/notes" "POST" $LIGHT_LOAD_USERS 30 "$note_data" "true"
    
    # Test public notes (no auth)
    make_concurrent_requests "/api/v1/notes/public?page=1&page_size=10" "GET" $MEDIUM_LOAD_USERS $TEST_DURATION "" "false"
    
    # Test search
    local search_data='{"query":"test","page":1,"page_size":10}'
    make_concurrent_requests "/api/v1/notes/search" "POST" $LIGHT_LOAD_USERS 30 "$search_data" "true"
}

# Function to run session management performance test
test_session_performance() {
    print_status "TEST" "Testing Session Management Performance"
    log_to_report "=== Session Management Performance ==="
    
    # Test active sessions
    make_concurrent_requests "/api/v1/user/sessions/active" "GET" $LIGHT_LOAD_USERS $TEST_DURATION "" "true"
    make_concurrent_requests "/api/v1/user/sessions/active" "GET" $MEDIUM_LOAD_USERS $TEST_DURATION "" "true"
    
    # Test session stats
    make_concurrent_requests "/api/v1/user/sessions/stats" "GET" $LIGHT_LOAD_USERS $TEST_DURATION "" "true"
}

# Function to run stress testing
test_stress_performance() {
    print_status "TEST" "Running Stress Tests"
    log_to_report "=== Stress Testing ==="
    
    print_status "WARNING" "Starting stress test with $STRESS_TEST_USERS concurrent users"
    
    # Stress test health endpoint
    make_concurrent_requests "/health" "GET" $STRESS_TEST_USERS 30 "" "false"
    
    # Stress test authenticated endpoints
    make_concurrent_requests "/api/v1/user/profile" "GET" $STRESS_TEST_USERS 30 "" "true"
    make_concurrent_requests "/api/v1/notes?page=1&page_size=10" "GET" $STRESS_TEST_USERS 30 "" "true"
}

# Function to run spike testing
test_spike_performance() {
    print_status "TEST" "Running Spike Tests"
    log_to_report "=== Spike Testing ==="
    
    print_status "WARNING" "Starting spike test with $SPIKE_TEST_USERS concurrent users for 15 seconds"
    
    # Spike test - sudden high load
    make_concurrent_requests "/health" "GET" $SPIKE_TEST_USERS 15 "" "false"
    make_concurrent_requests "/api/v1/user/profile" "GET" $SPIKE_TEST_USERS 15 "" "true"
}

# Function to run endurance testing
test_endurance_performance() {
    print_status "TEST" "Running Endurance Tests"
    log_to_report "=== Endurance Testing ==="
    
    print_status "INFO" "Running endurance test for 5 minutes with moderate load"
    
    # Endurance test - sustained load
    make_concurrent_requests "/health" "GET" $MEDIUM_LOAD_USERS 300 "" "false"
    make_concurrent_requests "/api/v1/user/profile" "GET" $LIGHT_LOAD_USERS 300 "" "true"
}

# Function to run memory usage monitoring
monitor_memory_usage() {
    print_status "INFO" "Monitoring system resources..."
    
    echo "System Resources Before Testing:" >> "$REPORT_FILE"
    echo "Memory Usage:" >> "$REPORT_FILE"
    free -h >> "$REPORT_FILE" 2>/dev/null || echo "Memory info not available" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    echo "Disk Usage:" >> "$REPORT_FILE"
    df -h >> "$REPORT_FILE" 2>/dev/null || echo "Disk info not available" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    if command -v ps &> /dev/null; then
        echo "GoNotes Process Info:" >> "$REPORT_FILE"
        ps aux | grep -E "(gonotes|go run)" | grep -v grep >> "$REPORT_FILE" 2>/dev/null || echo "No GoNotes process found" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    fi
}

# Function to generate summary report
generate_summary() {
    print_status "INFO" "Generating performance summary..."
    
    echo "" >> "$REPORT_FILE"
    echo "=== PERFORMANCE TESTING SUMMARY ===" >> "$REPORT_FILE"
    echo "Test completed at: $(date)" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # System resources after testing
    echo "System Resources After Testing:" >> "$REPORT_FILE"
    free -h >> "$REPORT_FILE" 2>/dev/null || echo "Memory info not available" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    echo "Report saved to: $REPORT_FILE" >> "$REPORT_FILE"
    
    print_status "SUCCESS" "Performance testing completed!"
    print_status "INFO" "Detailed report saved to: $REPORT_FILE"
}

# Function to cleanup
cleanup_test_data() {
    print_status "INFO" "Cleaning up test data..."
    
    # Logout to invalidate tokens
    if [[ -n "$REFRESH_TOKEN" ]]; then
        curl -s -X POST \
            -H "Content-Type: application/json" \
            -d '{"refresh_token":"'$REFRESH_TOKEN'"}' \
            "$BASE_URL/api/v1/auth/logout" > /dev/null 2>&1 || true
    fi
    
    # Clean up temporary files
    rm -f "$TEST_RESULTS_DIR"/temp_results_*.txt 2>/dev/null || true
    
    print_status "SUCCESS" "Cleanup completed"
}

# Main execution function
main() {
    print_status "INFO" "Starting GoNotes Performance Testing Suite"
    
    # Check dependencies
    if ! command -v bc &> /dev/null; then
        print_status "WARNING" "bc not found. Some calculations may be inaccurate"
    fi
    
    if ! command -v curl &> /dev/null; then
        print_status "ERROR" "curl is required for performance testing"
        exit 1
    fi
    
    # Initialize
    init_report
    check_server
    monitor_memory_usage
    
    # Setup test environment
    if ! setup_test_user; then
        print_status "ERROR" "Failed to setup test user. Exiting..."
        exit 1
    fi
    
    # Run performance tests
    print_status "INFO" "Running comprehensive performance tests..."
    
    test_health_performance
    test_auth_performance
    test_profile_performance
    test_notes_performance
    test_session_performance
    
    # Advanced testing
    print_status "INFO" "Running advanced performance tests..."
    test_stress_performance
    test_spike_performance
    
    # Optional: Endurance testing (uncomment if needed)
    # print_status "INFO" "Running endurance test (this may take several minutes)..."
    # test_endurance_performance
    
    # Generate final report
    generate_summary
    
    # Cleanup
    cleanup_test_data
    
    # Display final results
    echo ""
    print_status "SUCCESS" "🎉 Performance testing completed successfully!"
    print_status "INFO" "📊 Results saved to: $REPORT_FILE"
    print_status "INFO" "📁 All results in: $TEST_RESULTS_DIR/"
    
    if [[ -f "$REPORT_FILE" ]]; then
        echo ""
        print_status "INFO" "📋 Quick Summary:"
        tail -n 20 "$REPORT_FILE" | head -n 10
    fi
}

# Handle script arguments
case "${1:-}" in
    "--quick")
        LIGHT_LOAD_USERS=5
        MEDIUM_LOAD_USERS=15
        HEAVY_LOAD_USERS=25
        TEST_DURATION=30
        print_status "INFO" "Running in quick mode"
        ;;
    "--stress")
        STRESS_TEST_USERS=500
        SPIKE_TEST_USERS=1000
        print_status "INFO" "Running in high-stress mode"
        ;;
    "--help")
        echo "GoNotes Performance Testing Script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --quick    Run quick performance test (reduced load)"
        echo "  --stress   Run high-stress performance test"
        echo "  --help     Show this help message"
        echo ""
        echo "Default: Run standard performance test suite"
        exit 0
        ;;
esac

# Execute main function
main "$@" 