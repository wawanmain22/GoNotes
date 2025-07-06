#!/bin/bash

# Simple Integration Test Script
# Tests basic functionality of the GoNotes API

set -e

echo "🚀 GoNotes Simple Integration Test"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
BASE_URL="http://localhost:8080"
TEST_EMAIL="integration@test.com"
TEST_PASSWORD="testpassword123"
TEST_FULLNAME="Integration Test User"

# Global variables
ACCESS_TOKEN=""
REFRESH_TOKEN=""
NOTE_ID=""

# Test counters
PASSED=0
FAILED=0

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

# Function to run a test
run_test() {
    local test_name=$1
    local test_function=$2
    
    echo ""
    echo "Running: $test_name"
    echo "----------------------------------------"
    
    if $test_function; then
        print_status "SUCCESS" "$test_name PASSED"
        PASSED=$((PASSED + 1))
    else
        print_status "ERROR" "$test_name FAILED"
        FAILED=$((FAILED + 1))
    fi
}

# Test 1: Health Check
test_health_check() {
    local response=$(curl -s "$BASE_URL/health")
    
    if [[ $response == *"healthy"* ]]; then
        print_status "INFO" "Health check passed"
        return 0
    else
        print_status "ERROR" "Health check failed: $response"
        return 1
    fi
}

# Test 2: User Registration
test_user_registration() {
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$TEST_EMAIL\",
            \"password\": \"$TEST_PASSWORD\",
            \"full_name\": \"$TEST_FULLNAME\"
        }" \
        "$BASE_URL/api/v1/auth/register")
    
    if [[ $response == *"success"* ]]; then
        print_status "INFO" "User registration successful"
        return 0
    else
        print_status "ERROR" "Registration failed: $response"
        return 1
    fi
}

# Test 3: User Login
test_user_login() {
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$TEST_EMAIL\",
            \"password\": \"$TEST_PASSWORD\"
        }" \
        "$BASE_URL/api/v1/auth/login")
    
    if [[ $response == *"success"* ]]; then
        # Extract tokens (simple grep approach)
        ACCESS_TOKEN=$(echo "$response" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
        REFRESH_TOKEN=$(echo "$response" | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)
        
        if [[ -n "$ACCESS_TOKEN" && -n "$REFRESH_TOKEN" ]]; then
            print_status "INFO" "Login successful, tokens extracted"
            return 0
        else
            print_status "ERROR" "Login successful but failed to extract tokens"
            return 1
        fi
    else
        print_status "ERROR" "Login failed: $response"
        return 1
    fi
}

# Test 4: Get Profile
test_get_profile() {
    if [[ -z "$ACCESS_TOKEN" ]]; then
        print_status "ERROR" "No access token available"
        return 1
    fi
    
    local response=$(curl -s -X GET \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        "$BASE_URL/api/v1/user/profile")
    
    if [[ $response == *"success"* && $response == *"$TEST_EMAIL"* ]]; then
        print_status "INFO" "Profile retrieval successful"
        return 0
    else
        print_status "ERROR" "Profile retrieval failed: $response"
        return 1
    fi
}

# Test 5: Create Note
test_create_note() {
    if [[ -z "$ACCESS_TOKEN" ]]; then
        print_status "ERROR" "No access token available"
        return 1
    fi
    
    local response=$(curl -s -X POST \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"title\": \"Integration Test Note\",
            \"content\": \"This is a test note for integration testing\",
            \"tags\": [\"integration\", \"test\"]
        }" \
        "$BASE_URL/api/v1/notes")
    
    if [[ $response == *"success"* ]]; then
        # Extract note ID
        NOTE_ID=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
        
        if [[ -n "$NOTE_ID" ]]; then
            print_status "INFO" "Note creation successful, ID: $NOTE_ID"
            return 0
        else
            print_status "ERROR" "Note creation successful but failed to extract ID"
            return 1
        fi
    else
        print_status "ERROR" "Note creation failed: $response"
        return 1
    fi
}

# Test 6: Get Notes
test_get_notes() {
    if [[ -z "$ACCESS_TOKEN" ]]; then
        print_status "ERROR" "No access token available"
        return 1
    fi
    
    local response=$(curl -s -X GET \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        "$BASE_URL/api/v1/notes")
    
    if [[ $response == *"success"* ]]; then
        print_status "INFO" "Notes retrieval successful"
        return 0
    else
        print_status "ERROR" "Notes retrieval failed: $response"
        return 1
    fi
}

# Test 7: Token Refresh
test_token_refresh() {
    if [[ -z "$REFRESH_TOKEN" ]]; then
        print_status "ERROR" "No refresh token available"
        return 1
    fi
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"refresh_token\": \"$REFRESH_TOKEN\"
        }" \
        "$BASE_URL/api/v1/auth/refresh")
    
    if [[ $response == *"success"* ]]; then
        print_status "INFO" "Token refresh successful"
        return 0
    else
        print_status "ERROR" "Token refresh failed: $response"
        return 1
    fi
}

# Test 8: Logout
test_logout() {
    if [[ -z "$REFRESH_TOKEN" ]]; then
        print_status "ERROR" "No refresh token available"
        return 1
    fi
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"refresh_token\": \"$REFRESH_TOKEN\"
        }" \
        "$BASE_URL/api/v1/auth/logout")
    
    if [[ $response == *"success"* ]]; then
        print_status "INFO" "Logout successful"
        return 0
    else
        print_status "ERROR" "Logout failed: $response"
        return 1
    fi
}

# Main test execution
main() {
    print_status "INFO" "Starting Simple Integration Tests..."
    
    # Check if API is accessible
    if ! curl -s -f "$BASE_URL/health" > /dev/null; then
        print_status "ERROR" "API server is not accessible at $BASE_URL"
        print_status "INFO" "Make sure the GoNotes server is running"
        exit 1
    fi
    
    # Run tests
    run_test "Health Check" test_health_check
    run_test "User Registration" test_user_registration
    run_test "User Login" test_user_login
    run_test "Get Profile" test_get_profile
    run_test "Create Note" test_create_note
    run_test "Get Notes" test_get_notes
    run_test "Token Refresh" test_token_refresh
    run_test "Logout" test_logout
    
    # Print summary
    echo ""
    echo "========================================="
    echo "Integration Test Results"
    echo "========================================="
    echo "Total tests: $((PASSED + FAILED))"
    echo "Passed: $PASSED"
    echo "Failed: $FAILED"
    
    if [[ $FAILED -eq 0 ]]; then
        print_status "SUCCESS" "All integration tests passed!"
        exit 0
    else
        print_status "ERROR" "Some integration tests failed!"
        exit 1
    fi
}

# Run main function
main "$@" 