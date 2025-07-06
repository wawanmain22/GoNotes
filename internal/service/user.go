package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gonotes/internal/model"
	"gonotes/internal/repository"
	"gonotes/internal/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// UserService handles business logic for users
type UserService struct {
	userRepo    *repository.UserRepository
	redisClient *redis.Client
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// NewUserServiceWithRedis creates a new user service with Redis caching
func NewUserServiceWithRedis(userRepo *repository.UserRepository, redisClient *redis.Client) *UserService {
	return &UserService{
		userRepo:    userRepo,
		redisClient: redisClient,
	}
}

// Register creates a new user account
func (s *UserService) Register(req *model.RegisterRequest) (*model.User, error) {
	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationError(err))
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

	// Save to database
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user
func (s *UserService) Login(req *model.LoginRequest) (*model.User, error) {
	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationError(err))
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

// GetByID retrieves a user by ID
func (s *UserService) GetByID(id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// UpdateProfile updates a user's profile information
func (s *UserService) UpdateProfile(userID uuid.UUID, req *model.UpdateProfileRequest) (*model.User, error) {
	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationError(err))
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)

	// Get current user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if email is being changed and if the new email already exists
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

	// Save to database
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Invalidate profile cache after update
	if s.redisClient != nil {
		if err := utils.InvalidateProfileCache(s.redisClient, userID.String()); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to invalidate profile cache: %v\n", err)
		}
	}

	return user, nil
}

// GetProfileWithCache retrieves user profile with Redis caching
func (s *UserService) GetProfileWithCache(userID uuid.UUID) (*model.UserResponse, error) {
	// Try to get from cache first
	if s.redisClient != nil {
		cachedProfile, err := utils.GetProfileCache(s.redisClient, userID.String())
		if err == nil && cachedProfile != "" {
			var profile model.UserResponse
			if err := json.Unmarshal([]byte(cachedProfile), &profile); err == nil {
				return &profile, nil
			}
		}
	}

	// Get from database
	user, err := s.GetByID(userID)
	if err != nil {
		return nil, err
	}

	profile := user.ToResponse()

	// Cache the profile
	if s.redisClient != nil {
		profileJSON, err := json.Marshal(profile)
		if err == nil {
			// Cache for 3 minutes (180 seconds)
			if err := utils.SetProfileCache(s.redisClient, userID.String(), profileJSON, 180*time.Second); err != nil {
				// Log error but don't fail the request
				fmt.Printf("Failed to cache profile: %v\n", err)
			}
		}
	}

	return profile, nil
}
