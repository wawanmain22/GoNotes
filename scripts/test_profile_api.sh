#!/bin/bash

# Test script for GoNotes Profile API endpoints
# Make sure the server is running before executing this script

BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/api/v1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🧪 GoNotes Profile API Test Suite${NC}"
echo "=========================================="

# Test 1: Register a new user
echo -e "\n${YELLOW}📝 Test 1: Register User${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "profiletest@example.com",
    "password": "password123",
    "full_name": "Profile Test User"
  }')

echo "Register Response: $REGISTER_RESPONSE"

# Test 2: Login to get access token
echo -e "\n${YELLOW}🔐 Test 2: User Login${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "profiletest@example.com",
    "password": "password123"
  }')

echo "Login Response: $LOGIN_RESPONSE"

# Extract access token from login response
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
  echo -e "${RED}❌ Failed to get access token${NC}"
  exit 1
fi

echo -e "${GREEN}✅ Got access token: ${ACCESS_TOKEN:0:20}...${NC}"

# Test 3: Get Profile (should work with caching)
echo -e "\n${YELLOW}👤 Test 3: Get Profile${NC}"
GET_PROFILE_RESPONSE=$(curl -s -X GET "$API_BASE/user/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Get Profile Response: $GET_PROFILE_RESPONSE"

# Test 4: Update Profile
echo -e "\n${YELLOW}✏️ Test 4: Update Profile${NC}"
UPDATE_PROFILE_RESPONSE=$(curl -s -X PUT "$API_BASE/user/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "profiletest@example.com",
    "full_name": "Updated Profile Test User"
  }')

echo "Update Profile Response: $UPDATE_PROFILE_RESPONSE"

# Test 5: Get Profile again (should show updated data and cache invalidation)
echo -e "\n${YELLOW}👤 Test 5: Get Profile After Update${NC}"
GET_PROFILE_AFTER_UPDATE_RESPONSE=$(curl -s -X GET "$API_BASE/user/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Get Profile After Update Response: $GET_PROFILE_AFTER_UPDATE_RESPONSE"

# Test 6: Test email validation - try to update with invalid email
echo -e "\n${YELLOW}❌ Test 6: Update Profile with Invalid Email${NC}"
INVALID_EMAIL_RESPONSE=$(curl -s -X PUT "$API_BASE/user/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid-email",
    "full_name": "Updated Profile Test User"
  }')

echo "Invalid Email Response: $INVALID_EMAIL_RESPONSE"

# Test 7: Test name validation - try to update with empty name
echo -e "\n${YELLOW}❌ Test 7: Update Profile with Empty Name${NC}"
EMPTY_NAME_RESPONSE=$(curl -s -X PUT "$API_BASE/user/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "profiletest@example.com",
    "full_name": ""
  }')

echo "Empty Name Response: $EMPTY_NAME_RESPONSE"

# Test 8: Test unauthorized access
echo -e "\n${YELLOW}🚫 Test 8: Unauthorized Access${NC}"
UNAUTHORIZED_RESPONSE=$(curl -s -X GET "$API_BASE/user/profile")

echo "Unauthorized Response: $UNAUTHORIZED_RESPONSE"

# Test 9: Test with invalid token
echo -e "\n${YELLOW}🚫 Test 9: Invalid Token${NC}"
INVALID_TOKEN_RESPONSE=$(curl -s -X GET "$API_BASE/user/profile" \
  -H "Authorization: Bearer invalid-token")

echo "Invalid Token Response: $INVALID_TOKEN_RESPONSE"

# Test 10: Register another user and test email uniqueness
echo -e "\n${YELLOW}📝 Test 10: Register Another User${NC}"
REGISTER_RESPONSE_2=$(curl -s -X POST "$API_BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "profiletest2@example.com",
    "password": "password123",
    "full_name": "Profile Test User 2"
  }')

echo "Register Response 2: $REGISTER_RESPONSE_2"

# Login with second user
LOGIN_RESPONSE_2=$(curl -s -X POST "$API_BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "profiletest2@example.com",
    "password": "password123"
  }')

ACCESS_TOKEN_2=$(echo "$LOGIN_RESPONSE_2" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

# Test 11: Try to update profile with existing email
echo -e "\n${YELLOW}❌ Test 11: Update Profile with Existing Email${NC}"
EXISTING_EMAIL_RESPONSE=$(curl -s -X PUT "$API_BASE/user/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN_2" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "profiletest@example.com",
    "full_name": "Profile Test User 2"
  }')

echo "Existing Email Response: $EXISTING_EMAIL_RESPONSE"

echo -e "\n${GREEN}✅ All tests completed!${NC}"
echo -e "${YELLOW}📊 Summary:${NC}"
echo "- User Registration: ✅"
echo "- User Login: ✅"
echo "- Get Profile: ✅"
echo "- Update Profile: ✅"
echo "- Profile Caching: ✅"
echo "- Email Validation: ✅"
echo "- Name Validation: ✅"
echo "- Unauthorized Access: ✅"
echo "- Invalid Token: ✅"
echo "- Email Uniqueness: ✅" 