#!/bin/bash

# GoNotes API Testing Script - Notes Endpoints
# Tests all notes-related functionality including CRUD, search, pagination, etc.

set -e

# Configuration
BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/api/v1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Global variables for tokens and user info
ACCESS_TOKEN=""
REFRESH_TOKEN=""
USER_ID=""
TEST_NOTE_ID=""
TEST_NOTE_ID_2=""

# Counter for tests
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Helper function to print colored output
print_status() {
    local status=$1
    local message=$2
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✓ PASS${NC}: $message"
        ((PASSED_TESTS++))
    elif [ "$status" = "FAIL" ]; then
        echo -e "${RED}✗ FAIL${NC}: $message"
        ((FAILED_TESTS++))
    elif [ "$status" = "INFO" ]; then
        echo -e "${BLUE}ℹ INFO${NC}: $message"
    elif [ "$status" = "WARN" ]; then
        echo -e "${YELLOW}⚠ WARN${NC}: $message"
    fi
    ((TOTAL_TESTS++))
}

# Function to make HTTP requests with error handling
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local headers=$4
    
    if [ -n "$data" ]; then
        if [ -n "$headers" ]; then
            curl -s -X "$method" "$url" -H "Content-Type: application/json" -H "$headers" -d "$data"
        else
            curl -s -X "$method" "$url" -H "Content-Type: application/json" -d "$data"
        fi
    else
        if [ -n "$headers" ]; then
            curl -s -X "$method" "$url" -H "$headers"
        else
            curl -s -X "$method" "$url"
        fi
    fi
}

# Function to extract JSON field
extract_json_field() {
    local json=$1
    local field=$2
    echo "$json" | grep -o "\"$field\":\"[^\"]*\"" | cut -d'"' -f4
}

# Function to check if response contains expected field
check_response_field() {
    local response=$1
    local field=$2
    local expected=$3
    local test_name=$4
    
    local actual=$(extract_json_field "$response" "$field")
    if [ "$actual" = "$expected" ]; then
        print_status "PASS" "$test_name"
        return 0
    else
        print_status "FAIL" "$test_name - Expected: $expected, Got: $actual"
        return 1
    fi
}

# Function to check HTTP status in response
check_http_status() {
    local response=$1
    local expected_code=$2
    local test_name=$3
    
    local status_code=$(echo "$response" | grep -o '"code":[0-9]*' | cut -d':' -f2)
    if [ "$status_code" = "$expected_code" ]; then
        print_status "PASS" "$test_name"
        return 0
    else
        print_status "FAIL" "$test_name - Expected status: $expected_code, Got: $status_code"
        echo "Response: $response"
        return 1
    fi
}

# Setup: Register and login user for testing
setup_user() {
    print_status "INFO" "Setting up test user..."
    
    # Register test user
    local register_data='{
        "email": "testuser@gonotes.test",
        "password": "TestPassword123!",
        "full_name": "Test User"
    }'
    
    local register_response=$(make_request "POST" "$API_BASE/auth/register" "$register_data")
    check_response_field "$register_response" "status" "success" "User registration"
    
    # Login to get tokens
    local login_data='{
        "email": "testuser@gonotes.test", 
        "password": "TestPassword123!"
    }'
    
    local login_response=$(make_request "POST" "$API_BASE/auth/login" "$login_data")
    check_response_field "$login_response" "status" "success" "User login"
    
    # Extract tokens
    ACCESS_TOKEN=$(extract_json_field "$login_response" "access_token")
    REFRESH_TOKEN=$(extract_json_field "$login_response" "refresh_token")
    USER_ID=$(echo "$login_response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$ACCESS_TOKEN" ]; then
        print_status "PASS" "Access token extracted"
    else
        print_status "FAIL" "Failed to extract access token"
        exit 1
    fi
}

# Test 1: Create Note
test_create_note() {
    print_status "INFO" "Testing note creation..."
    
    local note_data='{
        "title": "Test Note 1",
        "content": "This is a test note content with some sample text.",
        "tags": ["test", "sample", "api"],
        "status": "active",
        "is_public": false
    }'
    
    local response=$(make_request "POST" "$API_BASE/notes" "$note_data" "Authorization: Bearer $ACCESS_TOKEN")
    
    if check_http_status "$response" "201" "Create note - HTTP status"; then
        TEST_NOTE_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        check_response_field "$response" "status" "success" "Create note - Response status"
        check_response_field "$response" "title" "Test Note 1" "Create note - Title verification"
    fi
}

# Test 2: Create second note for list testing
test_create_second_note() {
    print_status "INFO" "Creating second test note..."
    
    local note_data='{
        "title": "Test Note 2 - Public",
        "content": "This is a public test note for sharing.",
        "tags": ["test", "public"],
        "status": "active", 
        "is_public": true
    }'
    
    local response=$(make_request "POST" "$API_BASE/notes" "$note_data" "Authorization: Bearer $ACCESS_TOKEN")
    
    if check_http_status "$response" "201" "Create second note - HTTP status"; then
        TEST_NOTE_ID_2=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        check_response_field "$response" "is_public" "true" "Create second note - Public status"
    fi
}

# Test 3: Get Single Note
test_get_note() {
    print_status "INFO" "Testing get single note..."
    
    local response=$(make_request "GET" "$API_BASE/notes/$TEST_NOTE_ID" "" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "200" "Get note - HTTP status"
    check_response_field "$response" "status" "success" "Get note - Response status"
    check_response_field "$response" "title" "Test Note 1" "Get note - Title verification"
}

# Test 4: Get Notes List
test_get_notes_list() {
    print_status "INFO" "Testing get notes list..."
    
    local response=$(make_request "GET" "$API_BASE/notes?page=1&page_size=10" "" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "200" "Get notes list - HTTP status"
    check_response_field "$response" "status" "success" "Get notes list - Response status"
    
    # Check if notes array contains data
    if echo "$response" | grep -q '"notes":\['; then
        print_status "PASS" "Get notes list - Contains notes array"
    else
        print_status "FAIL" "Get notes list - Missing notes array"
    fi
}

# Test 5: Update Note
test_update_note() {
    print_status "INFO" "Testing note update..."
    
    local update_data='{
        "title": "Updated Test Note 1",
        "content": "This content has been updated via API.",
        "tags": ["test", "updated", "api"]
    }'
    
    local response=$(make_request "PUT" "$API_BASE/notes/$TEST_NOTE_ID" "$update_data" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "200" "Update note - HTTP status"
    check_response_field "$response" "status" "success" "Update note - Response status"
    check_response_field "$response" "title" "Updated Test Note 1" "Update note - Title updated"
}

# Test 6: Search Notes
test_search_notes() {
    print_status "INFO" "Testing notes search..."
    
    local search_data='{
        "query": "updated",
        "include_content": true,
        "page": 1,
        "page_size": 10
    }'
    
    local response=$(make_request "POST" "$API_BASE/notes/search" "$search_data" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "200" "Search notes - HTTP status"
    check_response_field "$response" "status" "success" "Search notes - Response status"
}

# Test 7: Get Notes by Tag
test_get_notes_by_tag() {
    print_status "INFO" "Testing get notes by tag..."
    
    local response=$(make_request "GET" "$API_BASE/notes/tag/test?page=1&page_size=10" "" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "200" "Get notes by tag - HTTP status"
    check_response_field "$response" "status" "success" "Get notes by tag - Response status"
}

# Test 8: Get User Tags
test_get_user_tags() {
    print_status "INFO" "Testing get user tags..."
    
    local response=$(make_request "GET" "$API_BASE/notes/tags" "" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "200" "Get user tags - HTTP status"
    check_response_field "$response" "status" "success" "Get user tags - Response status"
    
    # Check if tags array exists
    if echo "$response" | grep -q '"tags":\['; then
        print_status "PASS" "Get user tags - Contains tags array"
    else
        print_status "FAIL" "Get user tags - Missing tags array"
    fi
}

# Test 9: Get Notes Statistics
test_get_notes_stats() {
    print_status "INFO" "Testing get notes statistics..."
    
    local response=$(make_request "GET" "$API_BASE/notes/stats" "" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "200" "Get notes stats - HTTP status"
    check_response_field "$response" "status" "success" "Get notes stats - Response status"
    
    # Check if stats contain expected fields
    if echo "$response" | grep -q '"total":' && echo "$response" | grep -q '"active":'; then
        print_status "PASS" "Get notes stats - Contains statistics data"
    else
        print_status "FAIL" "Get notes stats - Missing statistics data"
    fi
}

# Test 10: Toggle Note Public Status
test_toggle_public_status() {
    print_status "INFO" "Testing toggle note public status..."
    
    local response=$(make_request "POST" "$API_BASE/notes/$TEST_NOTE_ID/toggle-public" "" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "200" "Toggle public status - HTTP status"
    check_response_field "$response" "status" "success" "Toggle public status - Response status"
    
    # Verify the note is now public
    if echo "$response" | grep -q '"is_public":true'; then
        print_status "PASS" "Toggle public status - Note made public"
    else
        print_status "FAIL" "Toggle public status - Note not made public"
    fi
}

# Test 11: Get Public Notes (No Auth Required)
test_get_public_notes() {
    print_status "INFO" "Testing get public notes (no auth)..."
    
    local response=$(make_request "GET" "$API_BASE/notes/public?page=1&page_size=10")
    
    check_http_status "$response" "200" "Get public notes - HTTP status"
    check_response_field "$response" "status" "success" "Get public notes - Response status"
    
    # Check if public notes are returned
    if echo "$response" | grep -q '"notes":\['; then
        print_status "PASS" "Get public notes - Contains notes array"
    else
        print_status "FAIL" "Get public notes - Missing notes array"
    fi
}

# Test 12: Duplicate Note
test_duplicate_note() {
    print_status "INFO" "Testing note duplication..."
    
    local response=$(make_request "POST" "$API_BASE/notes/$TEST_NOTE_ID/duplicate" "" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "201" "Duplicate note - HTTP status"
    check_response_field "$response" "status" "success" "Duplicate note - Response status"
    
    # Check if duplicated note has "(Copy)" in title
    if echo "$response" | grep -q "Copy"; then
        print_status "PASS" "Duplicate note - Title contains 'Copy'"
    else
        print_status "FAIL" "Duplicate note - Title doesn't contain 'Copy'"
    fi
}

# Test 13: Bulk Operations
test_bulk_operations() {
    print_status "INFO" "Testing bulk operations..."
    
    local bulk_data="{
        \"note_ids\": [\"$TEST_NOTE_ID\", \"$TEST_NOTE_ID_2\"],
        \"operation\": \"update_status\",
        \"data\": {\"status\": \"draft\"}
    }"
    
    local response=$(make_request "POST" "$API_BASE/notes/bulk" "$bulk_data" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "200" "Bulk operations - HTTP status"
    check_response_field "$response" "status" "success" "Bulk operations - Response status"
}

# Test 14: Soft Delete Note
test_delete_note() {
    print_status "INFO" "Testing note deletion..."
    
    local response=$(make_request "DELETE" "$API_BASE/notes/$TEST_NOTE_ID" "" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "200" "Delete note - HTTP status"
    check_response_field "$response" "status" "success" "Delete note - Response status"
}

# Test 15: Restore Deleted Note
test_restore_note() {
    print_status "INFO" "Testing note restoration..."
    
    local response=$(make_request "POST" "$API_BASE/notes/$TEST_NOTE_ID/restore" "" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "200" "Restore note - HTTP status"
    check_response_field "$response" "status" "success" "Restore note - Response status"
}

# Test 16: Error Cases - Unauthorized Access
test_unauthorized_access() {
    print_status "INFO" "Testing unauthorized access..."
    
    local response=$(make_request "GET" "$API_BASE/notes")
    
    check_http_status "$response" "401" "Unauthorized access - HTTP status"
    check_response_field "$response" "status" "error" "Unauthorized access - Response status"
}

# Test 17: Error Cases - Invalid Note ID
test_invalid_note_id() {
    print_status "INFO" "Testing invalid note ID..."
    
    local response=$(make_request "GET" "$API_BASE/notes/invalid-uuid" "" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "400" "Invalid note ID - HTTP status"
    check_response_field "$response" "status" "error" "Invalid note ID - Response status"
}

# Test 18: Error Cases - Note Not Found
test_note_not_found() {
    print_status "INFO" "Testing note not found..."
    
    local fake_uuid="123e4567-e89b-12d3-a456-426614174000"
    local response=$(make_request "GET" "$API_BASE/notes/$fake_uuid" "" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "404" "Note not found - HTTP status"
    check_response_field "$response" "status" "error" "Note not found - Response status"
}

# Test 19: Error Cases - Invalid JSON
test_invalid_json() {
    print_status "INFO" "Testing invalid JSON..."
    
    local invalid_json='{"title": "Test", "content"'
    local response=$(make_request "POST" "$API_BASE/notes" "$invalid_json" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "400" "Invalid JSON - HTTP status"
    check_response_field "$response" "status" "error" "Invalid JSON - Response status"
}

# Test 20: Error Cases - Validation Errors
test_validation_errors() {
    print_status "INFO" "Testing validation errors..."
    
    local invalid_data='{
        "title": "",
        "content": "Valid content"
    }'
    
    local response=$(make_request "POST" "$API_BASE/notes" "$invalid_data" "Authorization: Bearer $ACCESS_TOKEN")
    
    check_http_status "$response" "400" "Validation errors - HTTP status"
    check_response_field "$response" "status" "error" "Validation errors - Response status"
}

# Cleanup function
cleanup() {
    print_status "INFO" "Cleaning up test data..."
    
    if [ -n "$TEST_NOTE_ID" ]; then
        make_request "DELETE" "$API_BASE/notes/$TEST_NOTE_ID/hard" "" "Authorization: Bearer $ACCESS_TOKEN" > /dev/null 2>&1
    fi
    
    if [ -n "$TEST_NOTE_ID_2" ]; then
        make_request "DELETE" "$API_BASE/notes/$TEST_NOTE_ID_2/hard" "" "Authorization: Bearer $ACCESS_TOKEN" > /dev/null 2>&1
    fi
}

# Main testing function
run_tests() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  GoNotes API Testing - Notes Endpoints${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    # Check if server is running
    if ! curl -s "$BASE_URL/health" > /dev/null; then
        print_status "FAIL" "Server is not running at $BASE_URL"
        exit 1
    fi
    
    print_status "PASS" "Server is running"
    echo ""
    
    # Run tests
    setup_user
    echo ""
    
    test_create_note
    test_create_second_note
    test_get_note
    test_get_notes_list
    test_update_note
    test_search_notes
    test_get_notes_by_tag
    test_get_user_tags
    test_get_notes_stats
    test_toggle_public_status
    test_get_public_notes
    test_duplicate_note
    test_bulk_operations
    test_delete_note
    test_restore_note
    echo ""
    
    # Error case tests
    print_status "INFO" "Running error case tests..."
    test_unauthorized_access
    test_invalid_note_id
    test_note_not_found
    test_invalid_json
    test_validation_errors
    echo ""
    
    # Cleanup
    cleanup
    
    # Final results
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}            TEST RESULTS${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Total Tests: ${TOTAL_TESTS}"
    echo -e "${GREEN}Passed: ${PASSED_TESTS}${NC}"
    echo -e "${RED}Failed: ${FAILED_TESTS}${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}❌ Some tests failed.${NC}"
        exit 1
    fi
}

# Run the tests
run_tests 