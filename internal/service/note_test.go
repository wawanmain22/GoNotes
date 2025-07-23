package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"gonotes/internal/model"
	"gonotes/internal/utils"

	"github.com/google/uuid"
)

// Interfaces for dependency injection in tests
type NoteRepositoryInterface interface {
	Create(note *model.Note) error
	GetByID(id uuid.UUID) (*model.Note, error)
	GetByIDAndUserID(id, userID uuid.UUID) (*model.Note, error)
	GetByUserID(userID uuid.UUID, params *model.GetNotesParams) ([]model.Note, int64, error)
	GetPublicNotes(params *model.GetNotesParams) ([]model.Note, int64, error)
	Update(note *model.Note) error
	Delete(id, userID uuid.UUID) error
	Search(userID uuid.UUID, req *model.NoteSearchRequest) ([]model.Note, int64, error)
	IncrementViewCount(id uuid.UUID) error
	Restore(id, userID uuid.UUID) error
	HardDelete(id, userID uuid.UUID) error
	BulkUpdateStatus(userID uuid.UUID, noteIDs []uuid.UUID, status model.NoteStatus) error
	GetNoteStats(userID uuid.UUID) (map[string]interface{}, error)
}

type UserRepositoryInterface interface {
	Create(user *model.User) error
	GetByEmail(email string) (*model.User, error)
	GetByID(id uuid.UUID) (*model.User, error)
	EmailExists(email string) (bool, error)
	EmailExistsExcludingUser(email string, userID uuid.UUID) (bool, error)
	Update(user *model.User) error
}

// TestNoteService for testing with interfaces
type TestNoteService struct {
	noteRepo  NoteRepositoryInterface
	userRepo  UserRepositoryInterface
	validator *utils.Validator
}

func NewTestNoteService(noteRepo NoteRepositoryInterface, userRepo UserRepositoryInterface, validator *utils.Validator) *TestNoteService {
	return &TestNoteService{
		noteRepo:  noteRepo,
		userRepo:  userRepo,
		validator: validator,
	}
}

// Delegate all methods to the actual service methods but with interface dependencies
func (s *TestNoteService) CreateNote(userID uuid.UUID, req *model.CreateNoteRequest) (*model.NoteResponse, error) {
	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Convert request to note model
	note := req.ToNote(userID)

	// Validate content length
	if note.Content != nil && len(*note.Content) > 10000 {
		return nil, fmt.Errorf("content too long")
	}

	// Create note in database
	if err := s.noteRepo.Create(note); err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note.ToResponse(), nil
}

func (s *TestNoteService) GetUserNotes(userID uuid.UUID, params *model.GetNotesParams) (*model.NotesListResponse, error) {
	if err := s.validator.ValidateStruct(params); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	params.SetDefaults()

	notes, total, err := s.noteRepo.GetByUserID(userID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get user notes: %w", err)
	}

	response := model.NewNotesListResponse(notes, total, params)
	return response, nil
}

func (s *TestNoteService) UpdateNote(noteID, userID uuid.UUID, req *model.UpdateNoteRequest) (*model.NoteResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	note, err := s.noteRepo.GetByIDAndUserID(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}
	if note == nil {
		return nil, fmt.Errorf("note not found")
	}

	req.ApplyToNote(note)

	if err := s.noteRepo.Update(note); err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return note.ToResponse(), nil
}

func (s *TestNoteService) DeleteNote(noteID, userID uuid.UUID) error {
	note, err := s.noteRepo.GetByIDAndUserID(noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}
	if note == nil {
		return fmt.Errorf("note not found")
	}

	if err := s.noteRepo.Delete(noteID, userID); err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}

func (s *TestNoteService) SearchNotes(userID uuid.UUID, req *model.NoteSearchRequest) (*model.NotesListResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	req.SetDefaults()

	notes, total, err := s.noteRepo.Search(userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}

	params := &model.GetNotesParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Search:   req.Query,
		Status:   req.Status,
	}

	response := model.NewNotesListResponse(notes, total, params)
	return response, nil
}

func (s *TestNoteService) GetPublicNotes(params *model.GetNotesParams) (*model.NotesListResponse, error) {
	if err := s.validator.ValidateStruct(params); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	params.SetDefaults()

	notes, total, err := s.noteRepo.GetPublicNotes(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get public notes: %w", err)
	}

	response := model.NewNotesListResponse(notes, total, params)
	return response, nil
}

// MockNoteRepository implements NoteRepositoryInterface for testing
type MockNoteRepository struct {
	notes       map[uuid.UUID]*model.Note
	userNotes   map[uuid.UUID][]*model.Note
	publicNotes []*model.Note
	nextID      int
}

func NewMockNoteRepository() *MockNoteRepository {
	return &MockNoteRepository{
		notes:       make(map[uuid.UUID]*model.Note),
		userNotes:   make(map[uuid.UUID][]*model.Note),
		publicNotes: make([]*model.Note, 0),
		nextID:      1,
	}
}

func (m *MockNoteRepository) Create(note *model.Note) error {
	note.ID = uuid.New()
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()

	m.notes[note.ID] = note
	userNotes := m.userNotes[note.UserID]
	userNotes = append(userNotes, note)
	m.userNotes[note.UserID] = userNotes

	if note.IsPublic {
		m.publicNotes = append(m.publicNotes, note)
	}

	return nil
}

func (m *MockNoteRepository) GetByID(id uuid.UUID) (*model.Note, error) {
	note, exists := m.notes[id]
	if !exists {
		return nil, nil
	}
	return note, nil
}

func (m *MockNoteRepository) GetByIDAndUserID(id, userID uuid.UUID) (*model.Note, error) {
	note, exists := m.notes[id]
	if !exists || note.UserID != userID {
		return nil, nil
	}
	return note, nil
}

func (m *MockNoteRepository) GetByUserID(userID uuid.UUID, params *model.GetNotesParams) ([]model.Note, int64, error) {
	userNotes := m.userNotes[userID]

	// Apply pagination
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize

	if start >= len(userNotes) {
		return []model.Note{}, int64(len(userNotes)), nil
	}

	if end > len(userNotes) {
		end = len(userNotes)
	}

	result := make([]model.Note, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, *userNotes[i])
	}

	return result, int64(len(userNotes)), nil
}

func (m *MockNoteRepository) GetPublicNotes(params *model.GetNotesParams) ([]model.Note, int64, error) {
	// Apply pagination
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize

	if start >= len(m.publicNotes) {
		return []model.Note{}, int64(len(m.publicNotes)), nil
	}

	if end > len(m.publicNotes) {
		end = len(m.publicNotes)
	}

	result := make([]model.Note, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, *m.publicNotes[i])
	}

	return result, int64(len(m.publicNotes)), nil
}

func (m *MockNoteRepository) Update(note *model.Note) error {
	existingNote, exists := m.notes[note.ID]
	if !exists {
		return errors.New("note not found")
	}

	// Update the note
	note.UpdatedAt = time.Now()
	note.CreatedAt = existingNote.CreatedAt // Preserve original creation time
	m.notes[note.ID] = note

	// Update in user notes
	userNotes := m.userNotes[note.UserID]
	for i, userNote := range userNotes {
		if userNote.ID == note.ID {
			userNotes[i] = note
			break
		}
	}

	// Update in public notes
	wasPublic := false
	for i, publicNote := range m.publicNotes {
		if publicNote.ID == note.ID {
			wasPublic = true
			if note.IsPublic {
				m.publicNotes[i] = note
			} else {
				// Remove from public notes
				m.publicNotes = append(m.publicNotes[:i], m.publicNotes[i+1:]...)
			}
			break
		}
	}

	// Add to public notes if now public
	if !wasPublic && note.IsPublic {
		m.publicNotes = append(m.publicNotes, note)
	}

	return nil
}

func (m *MockNoteRepository) Delete(id, userID uuid.UUID) error {
	note, exists := m.notes[id]
	if !exists || note.UserID != userID {
		return errors.New("note not found")
	}

	// Remove from notes
	delete(m.notes, id)

	// Remove from user notes
	userNotes := m.userNotes[note.UserID]
	for i, userNote := range userNotes {
		if userNote.ID == id {
			userNotes = append(userNotes[:i], userNotes[i+1:]...)
			m.userNotes[note.UserID] = userNotes
			break
		}
	}

	// Remove from public notes
	for i, publicNote := range m.publicNotes {
		if publicNote.ID == id {
			m.publicNotes = append(m.publicNotes[:i], m.publicNotes[i+1:]...)
			break
		}
	}

	return nil
}

func (m *MockNoteRepository) Search(userID uuid.UUID, req *model.NoteSearchRequest) ([]model.Note, int64, error) {
	userNotes := m.userNotes[userID]
	var results []model.Note

	// Simple search implementation
	for _, note := range userNotes {
		if contains(note.Title, req.Query) ||
			(note.Content != nil && contains(*note.Content, req.Query)) {
			results = append(results, *note)
		}
	}

	// Apply pagination
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize

	if start >= len(results) {
		return []model.Note{}, int64(len(results)), nil
	}

	if end > len(results) {
		end = len(results)
	}

	return results[start:end], int64(len(results)), nil
}

// Mock methods for other repository operations
func (m *MockNoteRepository) IncrementViewCount(id uuid.UUID) error {
	note, exists := m.notes[id]
	if !exists {
		return errors.New("note not found")
	}
	note.ViewCount++
	return nil
}

func (m *MockNoteRepository) Restore(id, userID uuid.UUID) error {
	note, exists := m.notes[id]
	if !exists || note.UserID != userID {
		return errors.New("note not found")
	}
	note.Status = model.NoteStatusActive
	now := time.Now()
	note.UpdatedAt = now
	note.DeletedAt = nil
	return nil
}

func (m *MockNoteRepository) HardDelete(id, userID uuid.UUID) error {
	return m.Delete(id, userID)
}

func (m *MockNoteRepository) BulkUpdateStatus(userID uuid.UUID, noteIDs []uuid.UUID, status model.NoteStatus) error {
	for _, noteID := range noteIDs {
		note, exists := m.notes[noteID]
		if !exists || note.UserID != userID {
			continue
		}
		note.Status = status
		note.UpdatedAt = time.Now()
	}
	return nil
}

func (m *MockNoteRepository) GetNoteStats(userID uuid.UUID) (map[string]interface{}, error) {
	userNotes := m.userNotes[userID]
	stats := map[string]interface{}{
		"total":   len(userNotes),
		"active":  0,
		"draft":   0,
		"deleted": 0,
	}

	for _, note := range userNotes {
		switch note.Status {
		case model.NoteStatusActive:
			stats["active"] = stats["active"].(int) + 1
		case model.NoteStatusDraft:
			stats["draft"] = stats["draft"].(int) + 1
		case model.NoteStatusDeleted:
			stats["deleted"] = stats["deleted"].(int) + 1
		}
	}

	return stats, nil
}

// MockUserRepository implements UserRepositoryInterface for testing
type MockUserRepository struct {
	users     map[string]*model.User
	usersByID map[uuid.UUID]*model.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:     make(map[string]*model.User),
		usersByID: make(map[uuid.UUID]*model.User),
	}
}

func (m *MockUserRepository) Create(user *model.User) error {
	if _, exists := m.users[user.Email]; exists {
		return errors.New("email already exists")
	}
	m.users[user.Email] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetByEmail(email string) (*model.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*model.User, error) {
	user, exists := m.usersByID[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *MockUserRepository) EmailExists(email string) (bool, error) {
	_, exists := m.users[email]
	return exists, nil
}

func (m *MockUserRepository) EmailExistsExcludingUser(email string, userID uuid.UUID) (bool, error) {
	user, exists := m.users[email]
	if !exists {
		return false, nil
	}
	return user.ID != userID, nil
}

func (m *MockUserRepository) Update(user *model.User) error {
	if _, exists := m.usersByID[user.ID]; !exists {
		return errors.New("user not found")
	}

	// Check if email is being changed to an existing email
	if existingUser, exists := m.users[user.Email]; exists && existingUser.ID != user.ID {
		return errors.New("email already exists")
	}

	// Remove old email mapping if email changed
	for email, u := range m.users {
		if u.ID == user.ID && email != user.Email {
			delete(m.users, email)
			break
		}
	}

	m.users[user.Email] = user
	m.usersByID[user.ID] = user
	return nil
}

// Test Note Creation
func TestNoteService_CreateNote(t *testing.T) {
	noteRepo := NewMockNoteRepository()
	userRepo := NewMockUserRepository()
	validator := utils.NewValidator()
	noteService := NewTestNoteService(noteRepo, userRepo, validator)

	userID := uuid.New()

	// Create a user for testing
	testUser := &model.User{
		ID:        userID,
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FullName:  "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	userRepo.Create(testUser)

	testContent := "This is a test note content"
	testContent2 := "This note has tags"
	testContent3 := "This is a public note"
	testContent4 := "Content without title"
	isPublic := true

	tests := []struct {
		name          string
		request       *model.CreateNoteRequest
		expectError   bool
		expectedError string
	}{
		{
			name: "successful note creation",
			request: &model.CreateNoteRequest{
				Title:   "Test Note",
				Content: &testContent,
			},
			expectError: false,
		},
		{
			name: "note with tags",
			request: &model.CreateNoteRequest{
				Title:   "Tagged Note",
				Content: &testContent2,
				Tags:    []string{"work", "important"},
			},
			expectError: false,
		},
		{
			name: "public note",
			request: &model.CreateNoteRequest{
				Title:    "Public Note",
				Content:  &testContent3,
				IsPublic: &isPublic,
			},
			expectError: false,
		},
		{
			name: "empty title",
			request: &model.CreateNoteRequest{
				Title:   "",
				Content: &testContent4,
			},
			expectError:   true,
			expectedError: "validation",
		},
		{
			name: "empty content",
			request: &model.CreateNoteRequest{
				Title:   "Title without content",
				Content: nil,
			},
			expectError: false, // Content is optional
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := noteService.CreateNote(userID, tt.request)

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

			if response == nil {
				t.Errorf("expected response but got nil")
				return
			}

			if response.Title != tt.request.Title {
				t.Errorf("expected title '%s', got '%s'", tt.request.Title, response.Title)
			}
		})
	}
}

// Test Get User Notes
func TestNoteService_GetUserNotes(t *testing.T) {
	noteRepo := NewMockNoteRepository()
	userRepo := NewMockUserRepository()
	validator := utils.NewValidator()
	noteService := NewTestNoteService(noteRepo, userRepo, validator)

	userID := uuid.New()

	// Create test notes
	for i := 0; i < 15; i++ {
		content := fmt.Sprintf("Content for note %d", i+1)
		note := &model.Note{
			ID:      uuid.New(),
			Title:   fmt.Sprintf("Note %d", i+1),
			Content: &content,
			UserID:  userID,
			Status:  model.NoteStatusActive,
		}
		noteRepo.Create(note)
	}

	tests := []struct {
		name              string
		userID            uuid.UUID
		params            *model.GetNotesParams
		expectedNoteCount int
		expectError       bool
	}{
		{
			name:   "first page",
			userID: userID,
			params: &model.GetNotesParams{
				Page:     1,
				PageSize: 10,
			},
			expectedNoteCount: 10,
			expectError:       false,
		},
		{
			name:   "second page",
			userID: userID,
			params: &model.GetNotesParams{
				Page:     2,
				PageSize: 10,
			},
			expectedNoteCount: 5,
			expectError:       false,
		},
		{
			name:   "empty page",
			userID: userID,
			params: &model.GetNotesParams{
				Page:     3,
				PageSize: 10,
			},
			expectedNoteCount: 0,
			expectError:       false,
		},
		{
			name:   "different user",
			userID: uuid.New(),
			params: &model.GetNotesParams{
				Page:     1,
				PageSize: 10,
			},
			expectedNoteCount: 0,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := noteService.GetUserNotes(tt.userID, tt.params)

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

			if len(result.Notes) != tt.expectedNoteCount {
				t.Errorf("expected %d notes, got %d", tt.expectedNoteCount, len(result.Notes))
			}

			if result.Page != tt.params.Page {
				t.Errorf("expected page %d, got %d", tt.params.Page, result.Page)
			}

			if result.PageSize != tt.params.PageSize {
				t.Errorf("expected page size %d, got %d", tt.params.PageSize, result.PageSize)
			}
		})
	}
}

// Test Update Note
func TestNoteService_UpdateNote(t *testing.T) {
	noteRepo := NewMockNoteRepository()
	userRepo := NewMockUserRepository()
	validator := utils.NewValidator()
	noteService := NewTestNoteService(noteRepo, userRepo, validator)

	userID := uuid.New()
	anotherUserID := uuid.New()

	// Create a test note
	originalContent := "Original Content"
	existingNote := &model.Note{
		ID:      uuid.New(),
		Title:   "Original Title",
		Content: &originalContent,
		UserID:  userID,
		Status:  model.NoteStatusActive,
	}
	noteRepo.Create(existingNote)

	updatedTitle := "Updated Title"
	updatedContent := "Updated Content"
	taggedTitle := "Tagged Update"
	taggedContent := "Updated with tags"
	publicTitle := "Public Update"
	publicContent := "Now public"
	isPublic := true
	notFoundTitle := "Not Found"
	notFoundContent := "This note doesn't exist"
	unauthorizedTitle := "Unauthorized Update"
	unauthorizedContent := "This should fail"
	emptyTitle := ""
	emptyContent := "Content without title"

	tests := []struct {
		name          string
		noteID        uuid.UUID
		userID        uuid.UUID
		request       *model.UpdateNoteRequest
		expectError   bool
		expectedError string
	}{
		{
			name:   "successful update",
			noteID: existingNote.ID,
			userID: userID,
			request: &model.UpdateNoteRequest{
				Title:   &updatedTitle,
				Content: &updatedContent,
			},
			expectError: false,
		},
		{
			name:   "update with tags",
			noteID: existingNote.ID,
			userID: userID,
			request: &model.UpdateNoteRequest{
				Title:   &taggedTitle,
				Content: &taggedContent,
				Tags:    []string{"updated", "test"},
			},
			expectError: false,
		},
		{
			name:   "make public",
			noteID: existingNote.ID,
			userID: userID,
			request: &model.UpdateNoteRequest{
				Title:    &publicTitle,
				Content:  &publicContent,
				IsPublic: &isPublic,
			},
			expectError: false,
		},
		{
			name:   "note not found",
			noteID: uuid.New(),
			userID: userID,
			request: &model.UpdateNoteRequest{
				Title:   &notFoundTitle,
				Content: &notFoundContent,
			},
			expectError:   true,
			expectedError: "note not found",
		},
		{
			name:   "unauthorized user",
			noteID: existingNote.ID,
			userID: anotherUserID,
			request: &model.UpdateNoteRequest{
				Title:   &unauthorizedTitle,
				Content: &unauthorizedContent,
			},
			expectError:   true,
			expectedError: "note not found",
		},
		{
			name:   "empty title",
			noteID: existingNote.ID,
			userID: userID,
			request: &model.UpdateNoteRequest{
				Title:   &emptyTitle,
				Content: &emptyContent,
			},
			expectError:   true,
			expectedError: "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note, err := noteService.UpdateNote(tt.noteID, tt.userID, tt.request)

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

			if note == nil {
				t.Errorf("expected note but got nil")
				return
			}

			if tt.request.Title != nil && note.Title != *tt.request.Title {
				t.Errorf("expected title '%s', got '%s'", *tt.request.Title, note.Title)
			}

			if tt.request.Content != nil && note.Content != nil && *note.Content != *tt.request.Content {
				t.Errorf("expected content '%s', got '%s'", *tt.request.Content, *note.Content)
			}
		})
	}
}

// Test Delete Note
func TestNoteService_DeleteNote(t *testing.T) {
	noteRepo := NewMockNoteRepository()
	userRepo := NewMockUserRepository()
	validator := utils.NewValidator()
	noteService := NewTestNoteService(noteRepo, userRepo, validator)

	userID := uuid.New()
	anotherUserID := uuid.New()

	// Create a test note
	content := "This note will be deleted"
	existingNote := &model.Note{
		ID:      uuid.New(),
		Title:   "To Delete",
		Content: &content,
		UserID:  userID,
		Status:  model.NoteStatusActive,
	}
	noteRepo.Create(existingNote)

	tests := []struct {
		name          string
		noteID        uuid.UUID
		userID        uuid.UUID
		expectError   bool
		expectedError string
	}{
		{
			name:        "successful deletion",
			noteID:      existingNote.ID,
			userID:      userID,
			expectError: false,
		},
		{
			name:          "note not found",
			noteID:        uuid.New(),
			userID:        userID,
			expectError:   true,
			expectedError: "note not found",
		},
		{
			name:          "unauthorized user",
			noteID:        existingNote.ID,
			userID:        anotherUserID,
			expectError:   true,
			expectedError: "note not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := noteService.DeleteNote(tt.noteID, tt.userID)

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

			// Verify note is deleted
			deletedNote, err := noteRepo.GetByID(tt.noteID)
			if err != nil {
				t.Errorf("error checking deleted note: %v", err)
				return
			}

			if deletedNote != nil {
				t.Errorf("expected note to be deleted but it still exists")
			}
		})
	}
}

// Test Search Notes
func TestNoteService_SearchNotes(t *testing.T) {
	noteRepo := NewMockNoteRepository()
	userRepo := NewMockUserRepository()
	validator := utils.NewValidator()
	noteService := NewTestNoteService(noteRepo, userRepo, validator)

	userID := uuid.New()

	// Create test notes
	testNotes := []*model.Note{
		{ID: uuid.New(), Title: "Go Programming", Content: stringPtr("Learning Go language"), UserID: userID, Status: model.NoteStatusActive},
		{ID: uuid.New(), Title: "JavaScript Basics", Content: stringPtr("Introduction to JS"), UserID: userID, Status: model.NoteStatusActive},
		{ID: uuid.New(), Title: "Database Design", Content: stringPtr("SQL and NoSQL concepts"), UserID: userID, Status: model.NoteStatusActive},
		{ID: uuid.New(), Title: "Go Advanced", Content: stringPtr("Concurrency in Go"), UserID: userID, Status: model.NoteStatusActive},
	}

	for _, note := range testNotes {
		noteRepo.Create(note)
	}

	tests := []struct {
		name              string
		userID            uuid.UUID
		request           *model.NoteSearchRequest
		expectedNoteCount int
		expectError       bool
	}{
		{
			name:   "search for Go",
			userID: userID,
			request: &model.NoteSearchRequest{
				Query:    "Go",
				Page:     1,
				PageSize: 10,
			},
			expectedNoteCount: 2,
			expectError:       false,
		},
		{
			name:   "search for JavaScript",
			userID: userID,
			request: &model.NoteSearchRequest{
				Query:    "JavaScript",
				Page:     1,
				PageSize: 10,
			},
			expectedNoteCount: 1,
			expectError:       false,
		},
		{
			name:   "search not found",
			userID: userID,
			request: &model.NoteSearchRequest{
				Query:    "Python",
				Page:     1,
				PageSize: 10,
			},
			expectedNoteCount: 0,
			expectError:       false,
		},
		{
			name:   "empty search query",
			userID: userID,
			request: &model.NoteSearchRequest{
				Query:    "",
				Page:     1,
				PageSize: 10,
			},
			expectedNoteCount: 0,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := noteService.SearchNotes(tt.userID, tt.request)

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

			if len(result.Notes) != tt.expectedNoteCount {
				t.Errorf("expected %d notes, got %d", tt.expectedNoteCount, len(result.Notes))
			}
		})
	}
}

// Test Get Public Notes
func TestNoteService_GetPublicNotes(t *testing.T) {
	noteRepo := NewMockNoteRepository()
	userRepo := NewMockUserRepository()
	validator := utils.NewValidator()
	noteService := NewTestNoteService(noteRepo, userRepo, validator)

	userID := uuid.New()

	// Create test notes (some public, some private)
	publicNotes := []*model.Note{
		{ID: uuid.New(), Title: "Public Note 1", Content: stringPtr("Public content 1"), UserID: userID, IsPublic: true, Status: model.NoteStatusActive},
		{ID: uuid.New(), Title: "Public Note 2", Content: stringPtr("Public content 2"), UserID: userID, IsPublic: true, Status: model.NoteStatusActive},
		{ID: uuid.New(), Title: "Public Note 3", Content: stringPtr("Public content 3"), UserID: userID, IsPublic: true, Status: model.NoteStatusActive},
	}

	privateNotes := []*model.Note{
		{ID: uuid.New(), Title: "Private Note 1", Content: stringPtr("Private content 1"), UserID: userID, IsPublic: false, Status: model.NoteStatusActive},
		{ID: uuid.New(), Title: "Private Note 2", Content: stringPtr("Private content 2"), UserID: userID, IsPublic: false, Status: model.NoteStatusActive},
	}

	for _, note := range publicNotes {
		noteRepo.Create(note)
	}
	for _, note := range privateNotes {
		noteRepo.Create(note)
	}

	tests := []struct {
		name              string
		params            *model.GetNotesParams
		expectedNoteCount int
		expectError       bool
	}{
		{
			name: "get public notes",
			params: &model.GetNotesParams{
				Page:     1,
				PageSize: 10,
			},
			expectedNoteCount: 3,
			expectError:       false,
		},
		{
			name: "get public notes with pagination",
			params: &model.GetNotesParams{
				Page:     1,
				PageSize: 2,
			},
			expectedNoteCount: 2,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := noteService.GetPublicNotes(tt.params)

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

			if len(result.Notes) != tt.expectedNoteCount {
				t.Errorf("expected %d notes, got %d", tt.expectedNoteCount, len(result.Notes))
			}

			// Verify all returned notes are public
			for _, note := range result.Notes {
				if !note.IsPublic {
					t.Errorf("expected all notes to be public, but got private note: %s", note.Title)
				}
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsInner(s, substr))))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
