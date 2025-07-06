#!/bin/bash

# GoNotes API Testing Script
# This script tests all API endpoints with curl commands

# Configuration
BASE_URL="http://localhost:8080"
TEST_EMAIL="test-$(date +%s)@example.com"
TEST_PASSWORD="password123"
TEST_FULL_NAME="Test User"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Global variables for tokens
ACCESS_TOKEN=""
REFRESH_TOKEN=""

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "INFO")
            echo -e "${BLUE}[INFO]${NC} $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message"
            ;;
        "WARNING")
            echo -e "${YELLOW}[WARNING]${NC} $message"
            ;;
    esac
}

# Function to print test header
print_test_header() {
    echo
    echo "=========================================="
    echo "Testing: $1"
    echo "=========================================="
}

# Function to check HTTP status code
check_status() {
    local expected=$1
    local actual=$2
    local test_name=$3
    
    if [ "$actual" -eq "$expected" ]; then
        print_status "SUCCESS" "$test_name - Status: $actual"
        return 0
    else
        print_status "ERROR" "$test_name - Expected: $expected, Got: $actual"
        return 1
    fi
}

# Function to extract value from JSON response
extract_json_value() {
    local json=$1
    local key=$2
    echo "$json" | grep -o "\"$key\":\"[^\"]*\"" | cut -d'"' -f4
}

# Function to test health check
test_health_check() {
    print_test_header "Health Check"
    
    local response
    local status_code
    
    response=$(curl -s -w "%{http_code}" "$BASE_URL/health")
    status_code=${response: -3}
    response_body=${response%???}
    
    check_status 200 "$status_code" "Health Check"
    
    if [ $? -eq 0 ]; then
        print_status "INFO" "Response: $response_body"
    fi
}

# Function to test user registration
test_register() {
    print_test_header "User Registration"
    
    local response
    local status_code
    
    response=$(curl -s -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$TEST_EMAIL\",
            \"password\": \"$TEST_PASSWORD\",
            \"full_name\": \"$TEST_FULL_NAME\"
        }" \
        "$BASE_URL/api/v1/auth/register")
    
    status_code=${response: -3}
    response_body=${response%???}
    
    check_status 201 "$status_code" "User Registration"
    
    if [ $? -eq 0 ]; then
        print_status "SUCCESS" "User registered successfully"
        print_status "INFO" "Email: $TEST_EMAIL"
    else
        print_status "ERROR" "Registration failed: $response_body"
        exit 1
    fi
}

# Function to test user login
test_login() {
    print_test_header "User Login"
    
    local response
    local status_code
    
    response=$(curl -s -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$TEST_EMAIL\",
            \"password\": \"$TEST_PASSWORD\"
        }" \
        "$BASE_URL/api/v1/auth/login")
    
    status_code=${response: -3}
    response_body=${response%???}
    
    check_status 200 "$status_code" "User Login"
    
    if [ $? -eq 0 ]; then
        # Extract tokens from response
        ACCESS_TOKEN=$(extract_json_value "$response_body" "access_token")
        REFRESH_TOKEN=$(extract_json_value "$response_body" "refresh_token")
        
        if [ -n "$ACCESS_TOKEN" ] && [ -n "$REFRESH_TOKEN" ]; then
            print_status "SUCCESS" "Login successful, tokens obtained"
            print_status "INFO" "Access Token: ${ACCESS_TOKEN:0:50}..."
            print_status "INFO" "Refresh Token: ${REFRESH_TOKEN:0:50}..."
        else
            print_status "ERROR" "Failed to extract tokens from response"
            exit 1
        fi
    else
        print_status "ERROR" "Login failed: $response_body"
        exit 1
    fi
}

# Function to test get profile
test_get_profile() {
    print_test_header "Get User Profile"
    
    if [ -z "$ACCESS_TOKEN" ]; then
        print_status "ERROR" "No access token available"
        return 1
    fi
    
    local response
    local status_code
    
    response=$(curl -s -w "%{http_code}" \
        -X GET \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        "$BASE_URL/api/v1/user/profile")
    
    status_code=${response: -3}
    response_body=${response%???}
    
    check_status 200 "$status_code" "Get Profile"
    
    if [ $? -eq 0 ]; then
        print_status "SUCCESS" "Profile retrieved successfully"
        print_status "INFO" "Response: $response_body"
    else
        print_status "ERROR" "Get profile failed: $response_body"
    fi
}

# Function to test get sessions
test_get_sessions() {
    print_test_header "Get User Sessions"
    
    if [ -z "$ACCESS_TOKEN" ]; then
        print_status "ERROR" "No access token available"
        return 1
    fi
    
    local response
    local status_code
    
    response=$(curl -s -w "%{http_code}" \
        -X GET \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        "$BASE_URL/api/v1/user/sessions")
    
    status_code=${response: -3}
    response_body=${response%???}
    
    check_status 200 "$status_code" "Get Sessions"
    
    if [ $? -eq 0 ]; then
        print_status "SUCCESS" "Sessions retrieved successfully"
        print_status "INFO" "Response: $response_body"
    else
        print_status "ERROR" "Get sessions failed: $response_body"
    fi
}

# Function to test refresh token
test_refresh_token() {
    print_test_header "Refresh Token"
    
    if [ -z "$REFRESH_TOKEN" ]; then
        print_status "ERROR" "No refresh token available"
        return 1
    fi
    
    local response
    local status_code
    local old_access_token="$ACCESS_TOKEN"
    
    response=$(curl -s -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"refresh_token\": \"$REFRESH_TOKEN\"
        }" \
        "$BASE_URL/api/v1/auth/refresh")
    
    status_code=${response: -3}
    response_body=${response%???}
    
    check_status 200 "$status_code" "Refresh Token"
    
    if [ $? -eq 0 ]; then
        # Extract new access token
        NEW_ACCESS_TOKEN=$(extract_json_value "$response_body" "access_token")
        
        if [ -n "$NEW_ACCESS_TOKEN" ]; then
            ACCESS_TOKEN="$NEW_ACCESS_TOKEN"
            print_status "SUCCESS" "Token refreshed successfully"
            print_status "INFO" "New Access Token: ${ACCESS_TOKEN:0:50}..."
            
            # Verify new token is different
            if [ "$old_access_token" != "$ACCESS_TOKEN" ]; then
                print_status "SUCCESS" "New access token is different from old one"
            else
                print_status "WARNING" "New access token is same as old one"
            fi
        else
            print_status "ERROR" "Failed to extract new access token"
        fi
    else
        print_status "ERROR" "Refresh token failed: $response_body"
    fi
}

# Function to test notes endpoint (placeholder)
test_notes_endpoint() {
    print_test_header "Notes Endpoint (Placeholder)"
    
    if [ -z "$ACCESS_TOKEN" ]; then
        print_status "ERROR" "No access token available"
        return 1
    fi
    
    local response
    local status_code
    
    response=$(curl -s -w "%{http_code}" \
        -X GET \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        "$BASE_URL/api/v1/notes")
    
    status_code=${response: -3}
    response_body=${response%???}
    
    check_status 200 "$status_code" "Notes Endpoint"
    
    if [ $? -eq 0 ]; then
        print_status "SUCCESS" "Notes endpoint accessible"
        print_status "INFO" "Response: $response_body"
    else
        print_status "ERROR" "Notes endpoint failed: $response_body"
    fi
}

# Function to test logout
test_logout() {
    print_test_header "User Logout"
    
    if [ -z "$REFRESH_TOKEN" ]; then
        print_status "ERROR" "No refresh token available"
        return 1
    fi
    
    local response
    local status_code
    
    response=$(curl -s -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"refresh_token\": \"$REFRESH_TOKEN\"
        }" \
        "$BASE_URL/api/v1/auth/logout")
    
    status_code=${response: -3}
    response_body=${response%???}
    
    check_status 200 "$status_code" "User Logout"
    
    if [ $? -eq 0 ]; then
        print_status "SUCCESS" "Logout successful"
        print_status "INFO" "Response: $response_body"
    else
        print_status "ERROR" "Logout failed: $response_body"
    fi
}

# Function to test error scenarios
test_error_scenarios() {
    print_test_header "Error Scenarios"
    
    # Test invalid login
    print_status "INFO" "Testing invalid login credentials..."
    local response
    local status_code
    
    response=$(curl -s -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"nonexistent@example.com\",
            \"password\": \"wrongpassword\"
        }" \
        "$BASE_URL/api/v1/auth/login")
    
    status_code=${response: -3}
    response_body=${response%???}
    
    check_status 401 "$status_code" "Invalid Login"
    
    # Test unauthorized access
    print_status "INFO" "Testing unauthorized access..."
    response=$(curl -s -w "%{http_code}" \
        -X GET \
        "$BASE_URL/api/v1/user/profile")
    
    status_code=${response: -3}
    response_body=${response%???}
    
    check_status 401 "$status_code" "Unauthorized Access"
    
    # Test invalid refresh token
    print_status "INFO" "Testing invalid refresh token..."
    response=$(curl -s -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"refresh_token\": \"invalid_token\"
        }" \
        "$BASE_URL/api/v1/auth/refresh")
    
    status_code=${response: -3}
    response_body=${response%???}
    
    check_status 401 "$status_code" "Invalid Refresh Token"
}

# Function to verify logout invalidation
test_post_logout_invalidation() {
    print_test_header "Post-Logout Token Invalidation"
    
    # Try to use refresh token after logout (should fail)
    print_status "INFO" "Testing refresh token after logout..."
    local response
    local status_code
    
    response=$(curl -s -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"refresh_token\": \"$REFRESH_TOKEN\"
        }" \
        "$BASE_URL/api/v1/auth/refresh")
    
    status_code=${response: -3}
    response_body=${response%???}
    
    check_status 401 "$status_code" "Refresh Token After Logout"
    
    if [ $? -eq 0 ]; then
        print_status "SUCCESS" "Refresh token properly invalidated after logout"
    else
        print_status "ERROR" "Refresh token still valid after logout"
    fi
}

# Function to print summary
print_summary() {
    echo
    echo "=========================================="
    echo "TEST SUMMARY"
    echo "=========================================="
    print_status "INFO" "All tests completed"
    print_status "INFO" "Test Email: $TEST_EMAIL"
    print_status "INFO" "Base URL: $BASE_URL"
    echo
}

# Main function
main() {
    echo "=========================================="
    echo "GoNotes API Testing Script"
    echo "=========================================="
    echo "Base URL: $BASE_URL"
    echo "Test Email: $TEST_EMAIL"
    echo "=========================================="
    
    # Check if server is running
    print_status "INFO" "Checking server availability..."
    if ! curl -s "$BASE_URL/health" > /dev/null; then
        print_status "ERROR" "Server is not running at $BASE_URL"
        print_status "INFO" "Please start the GoNotes server and try again"
        exit 1
    fi
    
    print_status "SUCCESS" "Server is running"
    
    # Run all tests
    test_health_check
    test_register
    test_login
    test_get_profile
    test_get_sessions
    test_refresh_token
    test_notes_endpoint
    test_logout
    test_post_logout_invalidation
    test_error_scenarios
    
    print_summary
}

# Run main function
main "$@" 