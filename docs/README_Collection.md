# GoNotes API Collection Guide

## Overview

This folder contains comprehensive API documentation and testing collection for the complete GoNotes system, including authentication, profile management, advanced session management, and notes system with full CRUD operations.

## Files Included

- **`API_Documentation.md`** - Complete API documentation with examples
- **`GoNotes_API_Collection.postman_collection.json`** - Postman collection for testing
- **`GoNotes_Environment.postman_environment.json`** - Environment variables for Postman
- **`README_Collection.md`** - This guide

## Quick Start

### 1. Import to Postman

#### Import Collection
1. Open Postman
2. Click "Import" button
3. Select `GoNotes_API_Collection.postman_collection.json`
4. Click "Import"

#### Import Environment
1. Click the environment dropdown (top right)
2. Click "Import"
3. Select `GoNotes_Environment.postman_environment.json`
4. Click "Import"
5. Select "GoNotes API Environment" from dropdown

### 2. Setup Environment Variables

The environment comes with pre-configured variables:

| Variable | Default Value | Description |
|----------|---------------|-------------|
| `BASE_URL` | `http://localhost:8080` | API base URL |
| `ACCESS_TOKEN` | _(empty)_ | JWT access token (auto-filled) |
| `REFRESH_TOKEN` | _(empty)_ | JWT refresh token (auto-filled) |
| `TEST_EMAIL` | `test@example.com` | Test user email |
| `TEST_PASSWORD` | `password123` | Test user password |
| `TEST_FULL_NAME` | `Test User` | Test user full name |
| `NOTE_ID` | _(empty)_ | Note ID for testing (auto-filled) |
| `SESSION_ID` | _(empty)_ | Session ID for testing (auto-filled) |

**Note:** Access and refresh tokens are automatically populated when you run the login request.

### 3. Running the Collection

#### Manual Testing Sequence
1. **Start Server** - Ensure GoNotes server is running on `localhost:8080`
2. **Health Check** - Verify API is accessible
3. **Register User** - Create a test account
4. **Login** - Get authentication tokens
5. **Profile Management** - Test profile retrieval and updates
6. **Session Management** - Test active sessions and statistics
7. **Notes CRUD** - Create, read, update, delete notes
8. **Advanced Features** - Search notes, public notes, session management
9. **Refresh Token** - Test token refresh functionality
10. **Logout** - Test session invalidation and cleanup

#### Automated Testing
The collection includes test scripts that:
- Validate response status codes
- Check response structure
- Automatically extract and save tokens
- Verify error handling

### 4. Test Scripts Features

#### Automatic Token Management
```javascript
// Login request automatically saves tokens
pm.environment.set("ACCESS_TOKEN", jsonData.data.access_token);
pm.environment.set("REFRESH_TOKEN", jsonData.data.refresh_token);
```

#### Response Validation
```javascript
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response has required fields", function () {
    var jsonData = pm.response.json();
    pm.expect(jsonData.status).to.eql("success");
    pm.expect(jsonData.data).to.have.property("access_token");
});
```

## Collection Structure

### 1. Health Check
- **GET** `/health` - API health status

### 2. Authentication (4 requests)
- **POST** `/api/v1/auth/register` - User registration
- **POST** `/api/v1/auth/login` - User login
- **POST** `/api/v1/auth/refresh` - Token refresh
- **POST** `/api/v1/auth/logout` - User logout

### 3. User Management (2 requests)
- **GET** `/api/v1/user/profile` - Get user profile with Redis caching
- **PUT** `/api/v1/user/profile` - Update user profile with validation
- **GET** `/api/v1/user/sessions` - Get user sessions (legacy endpoint)

### 4. Advanced Session Management (5 requests)
- **GET** `/api/v1/user/sessions/active` - Get active sessions with device info
- **GET** `/api/v1/user/sessions/stats` - Get session statistics and analytics
- **DELETE** `/api/v1/user/sessions` - Invalidate all sessions (logout everywhere)
- **DELETE** `/api/v1/user/sessions/{sessionId}` - Invalidate specific session
- **POST** `/api/v1/user/sessions/invalidate` - Alternative session invalidation

### 5. Notes Management (7 requests)
- **GET** `/api/v1/notes` - Get user notes with pagination
- **POST** `/api/v1/notes` - Create new note with tags
- **GET** `/api/v1/notes/{id}` - Get specific note by ID
- **PUT** `/api/v1/notes/{id}` - Update existing note
- **POST** `/api/v1/notes/search` - Advanced search with filters
- **GET** `/api/v1/notes/public` - Get public notes (no auth required)
- **DELETE** `/api/v1/notes/{id}` - Delete note (soft delete)

### 6. Error Test Cases (9 requests)
- **Authentication Errors**: Invalid email registration, invalid credentials, invalid tokens
- **Profile Management Errors**: Invalid email format validation
- **Notes Errors**: Empty title validation, note not found
- **Session Errors**: Session not found, unauthorized access
- **Authorization Errors**: Missing tokens, expired tokens

## Running Tests

### Individual Request Testing
1. Select any request in the collection
2. Click "Send"
3. Check the "Test Results" tab for validation results

### Collection Runner
1. Right-click on "GoNotes API Collection"
2. Select "Run collection"
3. Configure settings:
   - **Iterations**: 1
   - **Delay**: 100ms between requests
   - **Data**: None (uses environment variables)
4. Click "Run GoNotes API Collection"

### Automated Test Flow
The collection is designed to run in sequence:
```
Health Check → Register → Login → Profile Management → 
Session Management → Notes CRUD → Advanced Features → 
Token Refresh → Logout & Cleanup
```

**Complete Flow (25+ requests):**
- Health Check (1)
- Authentication Flow (4)
- Profile Management (2) 
- Advanced Session Management (5)
- Notes Management (7)
- Error Testing (9)
- Token Refresh & Logout (2)

## Environment Configuration

### Development Environment
```json
{
  "BASE_URL": "http://localhost:8080",
  "TEST_EMAIL": "dev@example.com",
  "TEST_PASSWORD": "devpassword123"
}
```

### Staging Environment
```json
{
  "BASE_URL": "https://staging.gonotes.app",
  "TEST_EMAIL": "staging@example.com",
  "TEST_PASSWORD": "stagingpassword123"
}
```

### Production Environment
```json
{
  "BASE_URL": "https://api.gonotes.app",
  "TEST_EMAIL": "prod@example.com",
  "TEST_PASSWORD": "prodpassword123"
}
```

## Error Testing

### Common Error Scenarios
The collection includes tests for:

1. **400 Bad Request**
   - Invalid JSON format
   - Missing required fields
   - Validation errors

2. **401 Unauthorized**
   - Missing authorization header
   - Invalid credentials
   - Expired tokens

3. **500 Internal Server Error**
   - Server connectivity issues
   - Database connection problems

### Error Response Format
All errors follow the standard format:
```json
{
  "status": "error",
  "code": 400,
  "message": "Error description",
  "error": "Detailed error information"
}
```

## Comprehensive Testing Features

### Complete System Testing Cycle
1. **Authentication** - Register, login, token management
2. **Profile Management** - Get/update profile with caching
3. **Session Management** - Device tracking, multi-session handling
4. **Notes System** - CRUD operations, search, public sharing
5. **Advanced Features** - Pagination, filters, bulk operations
6. **Error Handling** - Comprehensive error scenario testing
7. **Security Testing** - Token validation, unauthorized access

### Authentication & Session Features
- **Access Token**: Valid for 15 minutes, stateless validation
- **Refresh Token**: Valid for 7 days with rotation
- **Session Tracking**: IP address and device detection
- **Multi-Device Support**: Login from multiple devices
- **Session Analytics**: Device breakdown and statistics
- **Secure Logout**: Single device or all devices

### Profile Management Features
- **Redis Caching**: 3-minute cache with automatic invalidation
- **Email Validation**: Comprehensive format and uniqueness checks
- **Profile Updates**: Real-time cache invalidation on changes
- **Data Consistency**: Atomic updates with rollback support

### Notes System Features
- **CRUD Operations**: Complete create, read, update, delete
- **Advanced Search**: Full-text search with tag filtering
- **Public Sharing**: Public/private note visibility
- **Pagination**: Efficient data loading with page controls
- **Tag System**: Multi-tag organization and filtering
- **Soft Delete**: Recovery support for accidentally deleted notes

## Tips & Best Practices

### 1. Environment Management
- Use different environments for different stages
- Keep sensitive data in environment variables
- Use meaningful variable names

### 2. Test Organization
- Group related requests in folders
- Use descriptive request names
- Add documentation to requests

### 3. Automated Testing
- Write comprehensive test scripts
- Validate both success and error cases
- Use environment variables for dynamic data

### 4. Error Handling
- Test edge cases and error scenarios
- Verify error response formats
- Check appropriate HTTP status codes

### 5. Security Testing
- Test unauthorized access attempts
- Verify token expiration handling
- Test session invalidation

## Troubleshooting

### Common Issues

#### 1. Connection Refused
- **Error**: `curl: (7) Failed to connect to localhost port 8080`
- **Solution**: Ensure GoNotes server is running

#### 2. Invalid Token
- **Error**: `"Invalid or expired token"`
- **Solution**: Re-run login request to get fresh tokens

#### 3. Environment Variables Not Set
- **Error**: Variables showing as `{{VARIABLE_NAME}}`
- **Solution**: Select correct environment in dropdown

#### 4. Test Failures
- **Error**: Test scripts failing
- **Solution**: Check server response format and update tests

### Debug Steps
1. Check server logs
2. Verify environment variables
3. Test individual requests
4. Check network connectivity
5. Validate JSON format

## Support

### Documentation
- Complete API documentation in `API_Documentation.md`
- Inline request documentation in Postman
- Error handling guide above

### Testing Help
- Use Postman Console for debugging
- Check test results after each request
- Use Collection Runner for automated testing

---

## Collection Summary

**Total Requests**: 25+ comprehensive API tests  
**Coverage**: Authentication, Profile Management, Session Management, Notes System  
**Test Scripts**: Automated validation with 50+ test assertions  
**Variables**: Auto-management of tokens, IDs, and test data  
**Error Testing**: 9 comprehensive error scenarios  

### Features Tested
✅ **Batch 1**: JWT Authentication with dual-token system  
✅ **Batch 2**: Complete Notes CRUD with search and public sharing  
✅ **Batch 3**: Profile Management with Redis caching and validation  
✅ **Batch 4**: Advanced Session Management with device tracking  

---

**Collection Version**: 2.0  
**Last Updated**: July 5, 2025  
**Compatible with**: Postman 10.0+  
**GoNotes API Version**: v1  
**Endpoints Covered**: 25+ requests across 6 categories 