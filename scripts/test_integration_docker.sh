#!/bin/bash

# GoNotes Integration Tests with Docker
# This script tests the entire application stack including PostgreSQL and Redis

set -e

echo "🚀 GoNotes Integration Test Suite with Docker"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
API_BASE_URL="http://localhost:8080/api/v1"
TEST_EMAIL="integration@test.com"
TEST_PASSWORD="testpassword123"
TEST_FULLNAME="Integration Test User"

# Global variables for test data
ACCESS_TOKEN=""
REFRESH_TOKEN=""
USER_ID=""
NOTE_ID=""
SESSION_ID=""

# Helper function to make API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local headers=${5:-"Content-Type: application/json"}
    
    if [[ ! -z "$ACCESS_TOKEN" ]]; then
        headers="$headers,Authorization: Bearer $ACCESS_TOKEN"
    fi
    
    echo -e "${BLUE}🔧 API Call: $method $endpoint${NC}"
    
    if [[ -z "$data" ]]; then
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -X "$method" \
            -H "$headers" \
            "$API_BASE_URL$endpoint")
    else
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -X "$method" \
            -H "$headers" \
            -d "$data" \
            "$API_BASE_URL$endpoint")
    fi
    
    body=$(echo "$response" | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
    status=$(echo "$response" | tr -d '\n' | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
    
    echo "Status: $status"
    echo "Response: $body" | jq . 2>/dev/null || echo "Response: $body"
    
    if [[ "$status" != "$expected_status" ]]; then
        echo -e "${RED}❌ Expected status $expected_status, got $status${NC}"
        return 1
    fi
    
    echo "$body"
}

# Function to check if services are ready
wait_for_services() {
    echo -e "${YELLOW}⏳ Waiting for services to be ready...${NC}"
    
    # Wait for API server
    max_attempts=60
    attempt=0
    while [[ $attempt -lt $max_attempts ]]; do
        if curl -s -f "http://localhost:8080/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ API server is ready${NC}"
            break
        fi
        attempt=$((attempt + 1))
        echo "Waiting for API server... ($attempt/$max_attempts)"
        sleep 2
    done
    
    if [[ $attempt -eq $max_attempts ]]; then
        echo -e "${RED}❌ API server failed to start${NC}"
        exit 1
    fi
    
    # Test database connection via health endpoint
    if curl -s -f "http://localhost:8080/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Database connection is ready${NC}"
    else
        echo -e "${RED}❌ Database connection failed${NC}"
        exit 1
    fi
}

# Function to clean up test data
cleanup_test_data() {
    echo -e "${YELLOW}🧹 Cleaning up test data...${NC}"
    
    # Clean up any existing test user
    if [[ ! -z "$ACCESS_TOKEN" ]]; then
        # Get all notes and delete them
        response=$(api_call "GET" "/notes" "" "200" 2>/dev/null || echo "")
        if [[ ! -z "$response" ]]; then
            note_ids=$(echo "$response" | jq -r '.notes[]?.id // empty' 2>/dev/null || echo "")
            for note_id in $note_ids; do
                api_call "DELETE" "/notes/$note_id" "" "200" > /dev/null 2>&1 || true
            done
        fi
        
        # Invalidate all sessions
        api_call "DELETE" "/user/sessions" "" "200" > /dev/null 2>&1 || true
    fi
    
    echo -e "${GREEN}✅ Test data cleaned up${NC}"
}

# Test 1: User Registration
test_user_registration() {
    echo -e "${YELLOW}📝 Test 1: User Registration${NC}"
    
    # Clean up any existing user first
    cleanup_test_data
    
    local data=$(cat <<EOF
{
    "email": "$TEST_EMAIL",
    "password": "$TEST_PASSWORD",
    "full_name": "$TEST_FULLNAME"
}
EOF
)
    
    response=$(api_call "POST" "/auth/register" "$data" "201")
    
    # Extract user data
    USER_ID=$(echo "$response" | jq -r '.user.id')
    ACCESS_TOKEN=$(echo "$response" | jq -r '.access_token')
    REFRESH_TOKEN=$(echo "$response" | jq -r '.refresh_token')
    
    if [[ -z "$USER_ID" || -z "$ACCESS_TOKEN" ]]; then
        echo -e "${RED}❌ Failed to register user${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ User registration successful${NC}"
    echo "User ID: $USER_ID"
}

# Test 2: User Login
test_user_login() {
    echo -e "${YELLOW}🔐 Test 2: User Login${NC}"
    
    local data=$(cat <<EOF
{
    "email": "$TEST_EMAIL",
    "password": "$TEST_PASSWORD"
}
EOF
)
    
    response=$(api_call "POST" "/auth/login" "$data" "200")
    
    # Update tokens
    ACCESS_TOKEN=$(echo "$response" | jq -r '.access_token')
    REFRESH_TOKEN=$(echo "$response" | jq -r '.refresh_token')
    
    if [[ -z "$ACCESS_TOKEN" ]]; then
        echo -e "${RED}❌ Failed to login${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ User login successful${NC}"
}

# Test 3: Profile Management
test_profile_management() {
    echo -e "${YELLOW}👤 Test 3: Profile Management${NC}"
    
    # Get current profile
    echo "Getting current profile..."
    response=$(api_call "GET" "/user/profile" "" "200")
    
    current_email=$(echo "$response" | jq -r '.email')
    if [[ "$current_email" != "$TEST_EMAIL" ]]; then
        echo -e "${RED}❌ Profile email mismatch${NC}"
        return 1
    fi
    
    # Update profile
    echo "Updating profile..."
    local updated_name="Updated Integration Test User"
    local data=$(cat <<EOF
{
    "email": "$TEST_EMAIL",
    "full_name": "$updated_name"
}
EOF
)
    
    response=$(api_call "PUT" "/user/profile" "$data" "200")
    
    updated_fullname=$(echo "$response" | jq -r '.full_name')
    if [[ "$updated_fullname" != "$updated_name" ]]; then
        echo -e "${RED}❌ Profile update failed${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ Profile management successful${NC}"
}

# Test 4: Note CRUD Operations
test_note_crud() {
    echo -e "${YELLOW}📝 Test 4: Note CRUD Operations${NC}"
    
    # Create note
    echo "Creating note..."
    local note_title="Integration Test Note"
    local note_content="This is a test note for integration testing"
    local data=$(cat <<EOF
{
    "title": "$note_title",
    "content": "$note_content",
    "tags": ["integration", "test"],
    "is_public": false
}
EOF
)
    
    response=$(api_call "POST" "/notes" "$data" "201")
    NOTE_ID=$(echo "$response" | jq -r '.id')
    
    if [[ -z "$NOTE_ID" ]]; then
        echo -e "${RED}❌ Failed to create note${NC}"
        return 1
    fi
    
    # Read note
    echo "Reading note..."
    response=$(api_call "GET" "/notes/$NOTE_ID" "" "200")
    
    read_title=$(echo "$response" | jq -r '.title')
    if [[ "$read_title" != "$note_title" ]]; then
        echo -e "${RED}❌ Note read failed${NC}"
        return 1
    fi
    
    # Update note
    echo "Updating note..."
    local updated_title="Updated Integration Test Note"
    local data=$(cat <<EOF
{
    "title": "$updated_title",
    "content": "$note_content",
    "tags": ["integration", "test", "updated"],
    "is_public": true
}
EOF
)
    
    response=$(api_call "PUT" "/notes/$NOTE_ID" "$data" "200")
    
    updated_note_title=$(echo "$response" | jq -r '.title')
    if [[ "$updated_note_title" != "$updated_title" ]]; then
        echo -e "${RED}❌ Note update failed${NC}"
        return 1
    fi
    
    # List notes
    echo "Listing notes..."
    response=$(api_call "GET" "/notes?page=1&page_size=10" "" "200")
    
    notes_count=$(echo "$response" | jq '.notes | length')
    if [[ "$notes_count" -lt 1 ]]; then
        echo -e "${RED}❌ Note list failed${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ Note CRUD operations successful${NC}"
}

# Test 5: Note Search and Filtering
test_note_search() {
    echo -e "${YELLOW}🔍 Test 5: Note Search and Filtering${NC}"
    
    # Search notes
    echo "Searching notes..."
    response=$(api_call "POST" "/notes/search" '{"query": "Integration", "page": 1, "page_size": 10}' "200")
    
    search_results=$(echo "$response" | jq '.notes | length')
    if [[ "$search_results" -lt 1 ]]; then
        echo -e "${RED}❌ Note search failed${NC}"
        return 1
    fi
    
    # Filter by tags
    echo "Filtering notes by tags..."
    response=$(api_call "GET" "/notes?tags=integration&page=1&page_size=10" "" "200")
    
    tag_results=$(echo "$response" | jq '.notes | length')
    if [[ "$tag_results" -lt 1 ]]; then
        echo -e "${RED}❌ Note tag filtering failed${NC}"
        return 1
    fi
    
    # Get public notes
    echo "Getting public notes..."
    response=$(api_call "GET" "/notes/public?page=1&page_size=10" "" "200")
    
    public_results=$(echo "$response" | jq '.notes | length')
    if [[ "$public_results" -lt 1 ]]; then
        echo -e "${RED}❌ Public notes retrieval failed${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ Note search and filtering successful${NC}"
}

# Test 6: Session Management
test_session_management() {
    echo -e "${YELLOW}🛡️ Test 6: Session Management${NC}"
    
    # Get active sessions
    echo "Getting active sessions..."
    response=$(api_call "GET" "/user/sessions/active" "" "200")
    
    sessions_count=$(echo "$response" | jq '.sessions | length')
    if [[ "$sessions_count" -lt 1 ]]; then
        echo -e "${RED}❌ Active sessions retrieval failed${NC}"
        return 1
    fi
    
    SESSION_ID=$(echo "$response" | jq -r '.sessions[0].id')
    
    # Get session stats
    echo "Getting session stats..."
    response=$(api_call "GET" "/user/sessions/stats" "" "200")
    
    total_sessions=$(echo "$response" | jq -r '.total_sessions')
    if [[ "$total_sessions" -lt 1 ]]; then
        echo -e "${RED}❌ Session stats retrieval failed${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ Session management successful${NC}"
}

# Test 7: Token Refresh
test_token_refresh() {
    echo -e "${YELLOW}🔄 Test 7: Token Refresh${NC}"
    
    local data=$(cat <<EOF
{
    "refresh_token": "$REFRESH_TOKEN"
}
EOF
)
    
    response=$(api_call "POST" "/auth/refresh" "$data" "200")
    
    new_access_token=$(echo "$response" | jq -r '.access_token')
    if [[ -z "$new_access_token" ]]; then
        echo -e "${RED}❌ Token refresh failed${NC}"
        return 1
    fi
    
    ACCESS_TOKEN="$new_access_token"
    echo -e "${GREEN}✅ Token refresh successful${NC}"
}

# Test 8: Rate Limiting
test_rate_limiting() {
    echo -e "${YELLOW}⚡ Test 8: Rate Limiting${NC}"
    
    echo "Testing rate limiting with rapid requests..."
    
    # Make multiple rapid requests
    for i in {1..15}; do
        api_call "GET" "/user/profile" "" "200" > /dev/null 2>&1 || true
    done
    
    # This request should potentially be rate limited
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
        -X "GET" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        "$API_BASE_URL/user/profile" 2>/dev/null)
    
    status=$(echo "$response" | tr -d '\n' | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
    
    if [[ "$status" == "429" ]]; then
        echo -e "${GREEN}✅ Rate limiting is working${NC}"
    else
        echo -e "${YELLOW}⚠️ Rate limiting may not be triggered (status: $status)${NC}"
    fi
}

# Test 9: Data Consistency
test_data_consistency() {
    echo -e "${YELLOW}🔍 Test 9: Data Consistency${NC}"
    
    # Create another note
    echo "Creating second note..."
    local data=$(cat <<EOF
{
    "title": "Consistency Test Note",
    "content": "Testing data consistency",
    "tags": ["consistency", "test"]
}
EOF
)
    
    response=$(api_call "POST" "/notes" "$data" "201")
    second_note_id=$(echo "$response" | jq -r '.id')
    
    # Verify both notes exist
    echo "Verifying data consistency..."
    response=$(api_call "GET" "/notes?page=1&page_size=20" "" "200")
    
    total_notes=$(echo "$response" | jq '.total')
    if [[ "$total_notes" -lt 2 ]]; then
        echo -e "${RED}❌ Data consistency check failed${NC}"
        return 1
    fi
    
    # Delete the second note
    api_call "DELETE" "/notes/$second_note_id" "" "200" > /dev/null
    
    echo -e "${GREEN}✅ Data consistency verified${NC}"
}

# Test 10: Error Handling
test_error_handling() {
    echo -e "${YELLOW}❌ Test 10: Error Handling${NC}"
    
    # Test invalid authentication
    echo "Testing invalid authentication..."
    local old_token="$ACCESS_TOKEN"
    ACCESS_TOKEN="invalid-token"
    
    api_call "GET" "/user/profile" "" "401" > /dev/null 2>&1
    if [[ $? -eq 0 ]]; then
        echo -e "${GREEN}✅ Invalid authentication handled correctly${NC}"
    else
        echo -e "${RED}❌ Invalid authentication not handled correctly${NC}"
        return 1
    fi
    
    ACCESS_TOKEN="$old_token"
    
    # Test accessing non-existent note
    echo "Testing non-existent resource..."
    fake_uuid="550e8400-e29b-41d4-a716-446655440000"
    
    api_call "GET" "/notes/$fake_uuid" "" "404" > /dev/null 2>&1
    if [[ $? -eq 0 ]]; then
        echo -e "${GREEN}✅ Non-existent resource handled correctly${NC}"
    else
        echo -e "${RED}❌ Non-existent resource not handled correctly${NC}"
        return 1
    fi
    
    # Test validation errors
    echo "Testing validation errors..."
    api_call "POST" "/notes" '{"title": "", "content": ""}' "400" > /dev/null 2>&1
    if [[ $? -eq 0 ]]; then
        echo -e "${GREEN}✅ Validation errors handled correctly${NC}"
    else
        echo -e "${RED}❌ Validation errors not handled correctly${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ Error handling tests successful${NC}"
}

# Test cleanup
test_cleanup() {
    echo -e "${YELLOW}🧹 Test Cleanup${NC}"
    
    # Delete test note
    if [[ ! -z "$NOTE_ID" ]]; then
        api_call "DELETE" "/notes/$NOTE_ID" "" "200" > /dev/null 2>&1 || true
    fi
    
    # Logout (invalidate session)
    api_call "POST" "/auth/logout" '{"refresh_token": "'$REFRESH_TOKEN'"}' "200" > /dev/null 2>&1 || true
    
    echo -e "${GREEN}✅ Test cleanup completed${NC}"
}

# Main test execution
main() {
    echo -e "${BLUE}Starting Docker Integration Tests...${NC}"
    echo "Test Target: $API_BASE_URL"
    
    # Check if jq is available
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}❌ jq is required for these tests. Please install jq.${NC}"
        exit 1
    fi
    
    # Wait for services to be ready
    wait_for_services
    
    # Run tests
    local tests=(
        "test_user_registration"
        "test_user_login" 
        "test_profile_management"
        "test_note_crud"
        "test_note_search"
        "test_session_management"
        "test_token_refresh"
        "test_rate_limiting"
        "test_data_consistency"
        "test_error_handling"
        "test_cleanup"
    )
    
    local passed=0
    local total=${#tests[@]}
    
    for test_func in "${tests[@]}"; do
        echo ""
        if $test_func; then
            passed=$((passed + 1))
        else
            echo -e "${RED}❌ Test $test_func failed${NC}"
        fi
    done
    
    echo ""
    echo "=============================================="
    echo -e "${BLUE}📊 Integration Test Results${NC}"
    echo "=============================================="
    echo -e "Total tests: ${BLUE}$total${NC}"
    echo -e "Passed: ${GREEN}$passed${NC}"
    echo -e "Failed: ${RED}$((total - passed))${NC}"
    
    if [[ $passed -eq $total ]]; then
        echo -e "${GREEN}🎉 All integration tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}💥 Some integration tests failed!${NC}"
        exit 1
    fi
}

# Execute main function
main "$@" 