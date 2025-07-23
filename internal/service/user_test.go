package service

import (
	"fmt"
	"testing"
	"time"

	"gonotes/internal/model"
	"gonotes/internal/utils"

	"github.com/google/uuid"
)

// TestUserService for testing with interfaces
type TestUserService struct {
	userRepo  UserRepositoryInterface
	validator *utils.Validator
}

func NewTestUserService(userRepo UserRepositoryInterface, validator *utils.Validator) *TestUserService {
	return &TestUserService{
		userRepo:  userRepo,
		validator: validator,
	}
}

// Delegate all methods to the actual service methods but with interface dependencies
func (s *TestUserService) Register(req *model.RegisterRequest) (*model.User, error) {
	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if email already exists
	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Password:  hashedPassword,
		FullName:  req.FullName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *TestUserService) Login(req *model.LoginRequest) (*model.User, error) {
	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := utils.VerifyPassword(user.Password, req.Password); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	return user, nil
}

func (s *TestUserService) GetByID(userID uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *TestUserService) UpdateProfile(userID uuid.UUID, req *model.UpdateProfileRequest) (*model.User, error) {
	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if email is being changed to an existing email
	if req.Email != user.Email {
		exists, err := s.userRepo.EmailExistsExcludingUser(req.Email, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check email existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("email already exists")
		}
	}

	// Update user fields
	user.Email = req.Email
	user.FullName = req.FullName
	user.UpdatedAt = time.Now()

	// Update in database
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// Test User Registration
func TestUserService_Register(t *testing.T) {
	userRepo := NewMockUserRepository()
	validator := utils.NewValidator()
	userService := NewTestUserService(userRepo, validator)

	tests := []struct {
		name          string
		request       *model.RegisterRequest
		expectError   bool
		expectedError string
	}{
		{
			name: "successful registration",
			request: &model.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				FullName: "Test User",
			},
			expectError: false,
		},
		{
			name: "invalid email",
			request: &model.RegisterRequest{
				Email:    "invalid-email",
				Password: "password123",
				FullName: "Test User",
			},
			expectError:   true,
			expectedError: "validation failed",
		},
		{
			name: "short password",
			request: &model.RegisterRequest{
				Email:    "test2@example.com",
				Password: "123",
				FullName: "Test User",
			},
			expectError:   true,
			expectedError: "validation failed",
		},
		{
			name: "short name",
			request: &model.RegisterRequest{
				Email:    "test3@example.com",
				Password: "password123",
				FullName: "T",
			},
			expectError:   true,
			expectedError: "validation failed",
		},
		{
			name: "duplicate email",
			request: &model.RegisterRequest{
				Email:    "test@example.com", // Same as first test
				Password: "password456",
				FullName: "Another User",
			},
			expectError:   true,
			expectedError: "email already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userService.Register(tt.request)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.expectedError != "" && !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Errorf("expected user but got nil")
				return
			}

			if user.Email != tt.request.Email {
				t.Errorf("expected email %s, got %s", tt.request.Email, user.Email)
			}

			if user.FullName != tt.request.FullName {
				t.Errorf("expected full name %s, got %s", tt.request.FullName, user.FullName)
			}

			// Verify password is hashed
			if user.Password == tt.request.Password {
				t.Errorf("password should be hashed")
			}

			// Verify password can be verified
			if err := utils.VerifyPassword(user.Password, tt.request.Password); err != nil {
				t.Errorf("password verification failed: %v", err)
			}
		})
	}
}

// Test User Login
func TestUserService_Login(t *testing.T) {
	userRepo := NewMockUserRepository()
	validator := utils.NewValidator()
	userService := NewTestUserService(userRepo, validator)

	// First register a user
	hashedPassword, _ := utils.HashPassword("password123")
	testUser := &model.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  hashedPassword,
		FullName:  "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	userRepo.Create(testUser)

	tests := []struct {
		name          string
		request       *model.LoginRequest
		expectError   bool
		expectedError string
	}{
		{
			name: "successful login",
			request: &model.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectError: false,
		},
		{
			name: "invalid email format",
			request: &model.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			expectError:   true,
			expectedError: "validation failed",
		},
		{
			name: "user not found",
			request: &model.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			expectError:   true,
			expectedError: "invalid email or password",
		},
		{
			name: "wrong password",
			request: &model.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			expectError:   true,
			expectedError: "invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userService.Login(tt.request)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.expectedError != "" && !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Errorf("expected user but got nil")
				return
			}

			if user.Email != tt.request.Email {
				t.Errorf("expected email %s, got %s", tt.request.Email, user.Email)
			}
		})
	}
}

// Test Update Profile
func TestUserService_UpdateProfile(t *testing.T) {
	userRepo := NewMockUserRepository()
	validator := utils.NewValidator()
	userService := NewTestUserService(userRepo, validator)

	// First register a user
	testUser := &model.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FullName:  "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	userRepo.Create(testUser)

	// Register another user for email conflict testing
	anotherUser := &model.User{
		ID:        uuid.New(),
		Email:     "another@example.com",
		Password:  "hashedpassword",
		FullName:  "Another User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	userRepo.Create(anotherUser)

	tests := []struct {
		name          string
		userID        uuid.UUID
		request       *model.UpdateProfileRequest
		expectError   bool
		expectedError string
	}{
		{
			name:   "successful update",
			userID: testUser.ID,
			request: &model.UpdateProfileRequest{
				Email:    "updated@example.com",
				FullName: "Updated User",
			},
			expectError: false,
		},
		{
			name:   "invalid email format",
			userID: testUser.ID,
			request: &model.UpdateProfileRequest{
				Email:    "invalid-email",
				FullName: "Updated User",
			},
			expectError:   true,
			expectedError: "validation failed",
		},
		{
			name:   "short name",
			userID: testUser.ID,
			request: &model.UpdateProfileRequest{
				Email:    "test2@example.com",
				FullName: "U",
			},
			expectError:   true,
			expectedError: "validation failed",
		},
		{
			name:   "email already exists",
			userID: testUser.ID,
			request: &model.UpdateProfileRequest{
				Email:    "another@example.com", // Another user's email
				FullName: "Updated User",
			},
			expectError:   true,
			expectedError: "email already exists",
		},
		{
			name:   "user not found",
			userID: uuid.New(), // Non-existent user
			request: &model.UpdateProfileRequest{
				Email:    "test3@example.com",
				FullName: "Updated User",
			},
			expectError:   true,
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userService.UpdateProfile(tt.userID, tt.request)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.expectedError != "" && !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Errorf("expected user but got nil")
				return
			}

			if user.Email != tt.request.Email {
				t.Errorf("expected email %s, got %s", tt.request.Email, user.Email)
			}

			if user.FullName != tt.request.FullName {
				t.Errorf("expected full name %s, got %s", tt.request.FullName, user.FullName)
			}
		})
	}
}

// Test Get By ID
func TestUserService_GetByID(t *testing.T) {
	userRepo := NewMockUserRepository()
	validator := utils.NewValidator()
	userService := NewTestUserService(userRepo, validator)

	// Register a test user
	testUser := &model.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FullName:  "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	userRepo.Create(testUser)

	tests := []struct {
		name          string
		userID        uuid.UUID
		expectError   bool
		expectedError string
	}{
		{
			name:        "user found",
			userID:      testUser.ID,
			expectError: false,
		},
		{
			name:          "user not found",
			userID:        uuid.New(),
			expectError:   true,
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userService.GetByID(tt.userID)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.expectedError != "" && !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Errorf("expected user but got nil")
				return
			}

			if user.ID != tt.userID {
				t.Errorf("expected user ID %s, got %s", tt.userID, user.ID)
			}
		})
	}
}
