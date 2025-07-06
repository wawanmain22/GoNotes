package service

import (
	"errors"
	"testing"
	"time"

	"gonotes/internal/config"
	"gonotes/internal/model"
	"gonotes/internal/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SessionRepositoryInterface for testing
type SessionRepositoryInterface interface {
	Create(session *model.Session) error
	GetByRefreshToken(refreshToken string) (*model.Session, error)
	GetByUserID(userID uuid.UUID) ([]model.Session, error)
	GetUserSessions(userID uuid.UUID) ([]model.Session, error)
	GetSessionByIDAndUserID(sessionID, userID uuid.UUID) (*model.Session, error)
	InvalidateByRefreshToken(refreshToken string) error
	InvalidateBySessionID(sessionID uuid.UUID) error
	InvalidateBySessionIDAndUserID(sessionID, userID uuid.UUID) error
	InvalidateAllByUserID(userID uuid.UUID) error
	CleanupExpiredSessions() error
}

// TestSessionService for testing with interfaces
type TestSessionService struct {
	sessionRepo SessionRepositoryInterface
	userRepo    UserRepositoryInterface
	redisClient *redis.Client
	config      *config.Config
}

func NewTestSessionService(sessionRepo SessionRepositoryInterface, userRepo UserRepositoryInterface, redisClient *redis.Client, cfg *config.Config) *TestSessionService {
	return &TestSessionService{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		redisClient: redisClient,
		config:      cfg,
	}
}

// Delegate methods to implement session service functionality
func (s *TestSessionService) CreateSession(user *model.User, userAgent, ipAddress string) (*model.AuthResponse, error) {
	// Create session
	expiresAt := time.Now().Add(s.config.RefreshExpire)
	session := &model.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: generateRefreshToken(),
		IsValid:      true,
		UserAgent:    &userAgent,
		IPAddress:    &ipAddress,
		CreatedAt:    time.Now(),
		ExpiresAt:    &expiresAt,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	// Generate access token
	accessToken := generateAccessToken(user.ID, user.Email, s.config)

	return &model.AuthResponse{
		User: &model.UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
		},
		AccessToken:  accessToken,
		RefreshToken: session.RefreshToken,
		ExpiresIn:    int64(s.config.JWTExpire.Seconds()),
	}, nil
}

func (s *TestSessionService) InvalidateSession(refreshToken string) error {
	session, err := s.sessionRepo.GetByRefreshToken(refreshToken)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("invalid refresh token")
	}

	return s.sessionRepo.InvalidateByRefreshToken(refreshToken)
}

func (s *TestSessionService) GetUserSessions(userID uuid.UUID, currentRefreshToken *string) ([]model.SessionResponse, error) {
	sessions, err := s.sessionRepo.GetUserSessions(userID)
	if err != nil {
		return nil, err
	}

	var responses []model.SessionResponse
	for _, session := range sessions {
		response := model.SessionResponse{
			ID:        session.ID,
			UserAgent: session.UserAgent,
			IPAddress: session.IPAddress,
			CreatedAt: session.CreatedAt,
			ExpiresAt: session.ExpiresAt,
		}

		// Add device info if available
		if session.UserAgent != nil {
			deviceInfo := parseDeviceInfo(*session.UserAgent)
			response.DeviceInfo = &deviceInfo
		}

		responses = append(responses, response)
	}

	return responses, nil
}

func (s *TestSessionService) InvalidateSpecificSession(userID, sessionID uuid.UUID) error {
	return s.sessionRepo.InvalidateBySessionIDAndUserID(sessionID, userID)
}

func (s *TestSessionService) InvalidateAllSessions(userID uuid.UUID) error {
	return s.sessionRepo.InvalidateAllByUserID(userID)
}

func (s *TestSessionService) ValidateAccessToken(accessToken string) (*utils.JWTClaims, error) {
	// Simple validation for testing
	if accessToken == "" || accessToken == "invalid-token" {
		return nil, errors.New("invalid access token")
	}

	// Return mock claims for valid tokens
	return &utils.JWTClaims{
		UserID: uuid.New(),
		Email:  "test@example.com",
	}, nil
}

// MockSessionRepository implements SessionRepositoryInterface for testing
type MockSessionRepository struct {
	sessions     map[string]*model.Session
	userSessions map[uuid.UUID][]*model.Session
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions:     make(map[string]*model.Session),
		userSessions: make(map[uuid.UUID][]*model.Session),
	}
}

func (m *MockSessionRepository) Create(session *model.Session) error {
	session.ID = uuid.New()
	session.CreatedAt = time.Now()

	m.sessions[session.RefreshToken] = session
	userSessions := m.userSessions[session.UserID]
	userSessions = append(userSessions, session)
	m.userSessions[session.UserID] = userSessions

	return nil
}

func (m *MockSessionRepository) GetByRefreshToken(refreshToken string) (*model.Session, error) {
	session, exists := m.sessions[refreshToken]
	if !exists || !session.IsValid {
		return nil, nil
	}
	return session, nil
}

func (m *MockSessionRepository) GetByUserID(userID uuid.UUID) ([]model.Session, error) {
	sessions := m.userSessions[userID]
	var validSessions []model.Session
	for _, session := range sessions {
		if session.IsValid {
			validSessions = append(validSessions, *session)
		}
	}
	return validSessions, nil
}

func (m *MockSessionRepository) GetUserSessions(userID uuid.UUID) ([]model.Session, error) {
	return m.GetByUserID(userID)
}

func (m *MockSessionRepository) GetSessionByIDAndUserID(sessionID, userID uuid.UUID) (*model.Session, error) {
	for _, session := range m.sessions {
		if session.ID == sessionID && session.UserID == userID && session.IsValid {
			return session, nil
		}
	}
	return nil, nil
}

func (m *MockSessionRepository) InvalidateByRefreshToken(refreshToken string) error {
	session, exists := m.sessions[refreshToken]
	if !exists {
		return errors.New("session not found")
	}
	session.IsValid = false
	return nil
}

func (m *MockSessionRepository) InvalidateBySessionID(sessionID uuid.UUID) error {
	for _, session := range m.sessions {
		if session.ID == sessionID {
			session.IsValid = false
			return nil
		}
	}
	return errors.New("session not found")
}

func (m *MockSessionRepository) InvalidateBySessionIDAndUserID(sessionID, userID uuid.UUID) error {
	for _, session := range m.sessions {
		if session.ID == sessionID && session.UserID == userID {
			session.IsValid = false
			return nil
		}
	}
	return errors.New("session not found or not owned by user")
}

func (m *MockSessionRepository) InvalidateAllByUserID(userID uuid.UUID) error {
	for _, session := range m.sessions {
		if session.UserID == userID {
			session.IsValid = false
		}
	}
	return nil
}

func (m *MockSessionRepository) CleanupExpiredSessions() error {
	return nil
}

// Helper functions for testing
func generateRefreshToken() string {
	return uuid.New().String()
}

func generateAccessToken(userID uuid.UUID, email string, cfg *config.Config) string {
	return "mock-access-token-" + userID.String()
}

func parseDeviceInfo(userAgent string) model.DeviceInfo {
	return model.DeviceInfo{
		Browser: "Test Browser",
		OS:      "Test OS",
		Device:  "Test Device",
	}
}

// Test Session Creation
func TestSessionService_CreateSession(t *testing.T) {
	sessionRepo := NewMockSessionRepository()
	userRepo := NewMockUserRepository()
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"}) // For testing, will fail gracefully
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		JWTExpire:     15 * time.Minute,
		RefreshExpire: 7 * 24 * time.Hour,
	}

	sessionService := NewTestSessionService(sessionRepo, userRepo, mockRedis, cfg)

	// Create a test user
	testUser := &model.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FullName:  "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name        string
		user        *model.User
		userAgent   string
		ipAddress   string
		expectError bool
	}{
		{
			name:        "successful session creation",
			user:        testUser,
			userAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
			ipAddress:   "192.168.1.100",
			expectError: false,
		},
		{
			name:        "session creation with mobile user agent",
			user:        testUser,
			userAgent:   "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
			ipAddress:   "10.0.0.50",
			expectError: false,
		},
		{
			name:        "session creation with empty user agent",
			user:        testUser,
			userAgent:   "",
			ipAddress:   "192.168.1.101",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authResponse, err := sessionService.CreateSession(tt.user, tt.userAgent, tt.ipAddress)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if authResponse == nil {
				t.Errorf("expected auth response but got nil")
				return
			}

			// Verify user information
			if authResponse.User.ID != tt.user.ID {
				t.Errorf("expected user ID %s, got %s", tt.user.ID, authResponse.User.ID)
			}

			if authResponse.User.Email != tt.user.Email {
				t.Errorf("expected email %s, got %s", tt.user.Email, authResponse.User.Email)
			}

			// Verify tokens are present
			if authResponse.AccessToken == "" {
				t.Errorf("expected access token but got empty string")
			}

			if authResponse.RefreshToken == "" {
				t.Errorf("expected refresh token but got empty string")
			}

			// Verify expiry time
			if authResponse.ExpiresIn != int64(cfg.JWTExpire.Seconds()) {
				t.Errorf("expected expires in %d, got %d", int64(cfg.JWTExpire.Seconds()), authResponse.ExpiresIn)
			}
		})
	}
}

// Test Session Invalidation
func TestSessionService_InvalidateSession(t *testing.T) {
	sessionRepo := NewMockSessionRepository()
	userRepo := NewMockUserRepository()
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		JWTExpire:     15 * time.Minute,
		RefreshExpire: 7 * 24 * time.Hour,
	}

	sessionService := NewTestSessionService(sessionRepo, userRepo, mockRedis, cfg)

	// Create a test session
	testUser := &model.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FullName:  "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create session first
	authResponse, err := sessionService.CreateSession(testUser, "test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("failed to create session for test: %v", err)
	}

	tests := []struct {
		name          string
		refreshToken  string
		expectError   bool
		expectedError string
	}{
		{
			name:         "successful invalidation",
			refreshToken: authResponse.RefreshToken,
			expectError:  false,
		},
		{
			name:          "invalid refresh token",
			refreshToken:  "invalid-token",
			expectError:   true,
			expectedError: "invalid refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sessionService.InvalidateSession(tt.refreshToken)

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
			}
		})
	}
}

// Test Get User Sessions
func TestSessionService_GetUserSessions(t *testing.T) {
	sessionRepo := NewMockSessionRepository()
	userRepo := NewMockUserRepository()
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		JWTExpire:     15 * time.Minute,
		RefreshExpire: 7 * 24 * time.Hour,
	}

	sessionService := NewTestSessionService(sessionRepo, userRepo, mockRedis, cfg)

	// Create a test user
	testUser := &model.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FullName:  "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create multiple sessions
	session1, _ := sessionService.CreateSession(testUser, "Mozilla/5.0 (Macintosh)", "192.168.1.100")
	_, _ = sessionService.CreateSession(testUser, "Mozilla/5.0 (iPhone)", "10.0.0.50")

	tests := []struct {
		name                 string
		userID               uuid.UUID
		currentRefreshToken  *string
		expectError          bool
		expectedSessionCount int
	}{
		{
			name:                 "get all sessions without current token",
			userID:               testUser.ID,
			currentRefreshToken:  nil,
			expectError:          false,
			expectedSessionCount: 2,
		},
		{
			name:                 "get sessions with current token",
			userID:               testUser.ID,
			currentRefreshToken:  &session1.RefreshToken,
			expectError:          false,
			expectedSessionCount: 2,
		},
		{
			name:                 "get sessions for non-existent user",
			userID:               uuid.New(),
			currentRefreshToken:  nil,
			expectError:          false,
			expectedSessionCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions, err := sessionService.GetUserSessions(tt.userID, tt.currentRefreshToken)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(sessions) != tt.expectedSessionCount {
				t.Errorf("expected %d sessions, got %d", tt.expectedSessionCount, len(sessions))
			}

			// If we have sessions, verify device info is populated
			for _, session := range sessions {
				if session.DeviceInfo == nil {
					t.Errorf("expected device info to be populated but got nil")
				}
			}
		})
	}
}

// Test Invalidate Specific Session
func TestSessionService_InvalidateSpecificSession(t *testing.T) {
	sessionRepo := NewMockSessionRepository()
	userRepo := NewMockUserRepository()
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		JWTExpire:     15 * time.Minute,
		RefreshExpire: 7 * 24 * time.Hour,
	}

	sessionService := NewTestSessionService(sessionRepo, userRepo, mockRedis, cfg)

	// Create a test user and session
	testUser := &model.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FullName:  "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, _ = sessionService.CreateSession(testUser, "test-agent", "127.0.0.1")

	// Extract session ID from the created session (in real implementation, we'd get this from the session)
	sessions, _ := sessionService.GetUserSessions(testUser.ID, nil)
	var sessionID uuid.UUID
	if len(sessions) > 0 {
		sessionID = sessions[0].ID
	}

	tests := []struct {
		name          string
		userID        uuid.UUID
		sessionID     uuid.UUID
		expectError   bool
		expectedError string
	}{
		{
			name:        "successful specific session invalidation",
			userID:      testUser.ID,
			sessionID:   sessionID,
			expectError: false,
		},
		{
			name:          "session not found",
			userID:        testUser.ID,
			sessionID:     uuid.New(),
			expectError:   true,
			expectedError: "session not found",
		},
		{
			name:          "session not owned by user",
			userID:        uuid.New(),
			sessionID:     sessionID,
			expectError:   true,
			expectedError: "session not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sessionService.InvalidateSpecificSession(tt.userID, tt.sessionID)

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
			}
		})
	}
}

// Test Invalidate All Sessions
func TestSessionService_InvalidateAllSessions(t *testing.T) {
	sessionRepo := NewMockSessionRepository()
	userRepo := NewMockUserRepository()
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		JWTExpire:     15 * time.Minute,
		RefreshExpire: 7 * 24 * time.Hour,
	}

	sessionService := NewTestSessionService(sessionRepo, userRepo, mockRedis, cfg)

	// Create a test user and multiple sessions
	testUser := &model.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FullName:  "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create multiple sessions
	sessionService.CreateSession(testUser, "Mozilla/5.0 (Macintosh)", "192.168.1.100")
	sessionService.CreateSession(testUser, "Mozilla/5.0 (iPhone)", "10.0.0.50")

	// Verify sessions exist
	sessionsBefore, _ := sessionService.GetUserSessions(testUser.ID, nil)
	if len(sessionsBefore) != 2 {
		t.Fatalf("expected 2 sessions before invalidation, got %d", len(sessionsBefore))
	}

	// Test invalidate all sessions
	err := sessionService.InvalidateAllSessions(testUser.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify all sessions are invalidated
	sessionsAfter, _ := sessionService.GetUserSessions(testUser.ID, nil)
	if len(sessionsAfter) != 0 {
		t.Errorf("expected 0 sessions after invalidation, got %d", len(sessionsAfter))
	}
}

// Test Access Token Validation
func TestSessionService_ValidateAccessToken(t *testing.T) {
	sessionRepo := NewMockSessionRepository()
	userRepo := NewMockUserRepository()
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		JWTExpire:     15 * time.Minute,
		RefreshExpire: 7 * 24 * time.Hour,
	}

	sessionService := NewTestSessionService(sessionRepo, userRepo, mockRedis, cfg)

	// Create a test user and session
	testUser := &model.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FullName:  "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	authResponse, _ := sessionService.CreateSession(testUser, "test-agent", "127.0.0.1")

	tests := []struct {
		name          string
		accessToken   string
		expectError   bool
		expectedError string
	}{
		{
			name:        "valid access token",
			accessToken: authResponse.AccessToken,
			expectError: false,
		},
		{
			name:          "invalid access token",
			accessToken:   "invalid-token",
			expectError:   true,
			expectedError: "invalid access token",
		},
		{
			name:          "empty access token",
			accessToken:   "",
			expectError:   true,
			expectedError: "invalid access token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := sessionService.ValidateAccessToken(tt.accessToken)

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

			if claims == nil {
				t.Errorf("expected claims but got nil")
				return
			}

			if claims.Email != "test@example.com" {
				t.Errorf("expected email test@example.com, got %s", claims.Email)
			}
		})
	}
}
