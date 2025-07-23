package service

import (
	"fmt"
	"strings"

	"gonotes/internal/model"
	"gonotes/internal/repository"
	"gonotes/internal/utils"

	"github.com/google/uuid"
)

// NoteService handles business logic for notes
type NoteService struct {
	noteRepo  *repository.NoteRepository
	userRepo  *repository.UserRepository
	validator *utils.Validator
}

// NewNoteService creates a new note service
func NewNoteService(noteRepo *repository.NoteRepository, userRepo *repository.UserRepository, validator *utils.Validator) *NoteService {
	return &NoteService{
		noteRepo:  noteRepo,
		userRepo:  userRepo,
		validator: validator,
	}
}

// CreateNote creates a new note
func (s *NoteService) CreateNote(userID uuid.UUID, req *model.CreateNoteRequest) (*model.NoteResponse, error) {
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

	// Validate and sanitize content
	if err := s.validateNoteContent(note); err != nil {
		return nil, err
	}

	// Create note in database
	if err := s.noteRepo.Create(note); err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	// Return response
	return note.ToResponse(), nil
}

// GetNoteByID retrieves a note by ID with security checks
func (s *NoteService) GetNoteByID(noteID, userID uuid.UUID) (*model.NoteResponse, error) {
	// Get note from database
	note, err := s.noteRepo.GetByID(noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}
	if note == nil {
		return nil, fmt.Errorf("note not found")
	}

	// Check if user can access this note
	if !s.canUserAccessNote(note, userID) {
		return nil, fmt.Errorf("access denied")
	}

	// Increment view count if it's not the owner viewing
	if note.UserID != userID {
		if err := s.noteRepo.IncrementViewCount(noteID); err != nil {
			// Log error but don't fail the request
			// In production, use proper logging
			fmt.Printf("Failed to increment view count: %v\n", err)
		}
	}

	return note.ToResponse(), nil
}

// GetUserNotes retrieves notes for a user with pagination and filtering
func (s *NoteService) GetUserNotes(userID uuid.UUID, params *model.GetNotesParams) (*model.NotesListResponse, error) {
	// Validate parameters
	if err := s.validator.ValidateStruct(params); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Set defaults
	params.SetDefaults()

	// Get notes from repository
	notes, total, err := s.noteRepo.GetByUserID(userID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get user notes: %w", err)
	}

	// Convert to response
	response := model.NewNotesListResponse(notes, total, params)
	return response, nil
}

// UpdateNote updates an existing note
func (s *NoteService) UpdateNote(noteID, userID uuid.UUID, req *model.UpdateNoteRequest) (*model.NoteResponse, error) {
	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Get existing note
	note, err := s.noteRepo.GetByIDAndUserID(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}
	if note == nil {
		return nil, fmt.Errorf("note not found")
	}

	// Check if note can be edited
	if !note.CanEdit() {
		return nil, fmt.Errorf("note cannot be edited in current status")
	}

	// Apply updates
	req.ApplyToNote(note)

	// Validate updated content
	if err := s.validateNoteContent(note); err != nil {
		return nil, err
	}

	// Update in database
	if err := s.noteRepo.Update(note); err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return note.ToResponse(), nil
}

// DeleteNote soft deletes a note
func (s *NoteService) DeleteNote(noteID, userID uuid.UUID) error {
	// Check if note exists and user has permission
	note, err := s.noteRepo.GetByIDAndUserID(noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}
	if note == nil {
		return fmt.Errorf("note not found")
	}

	// Check if note is already deleted
	if note.IsDeleted() {
		return fmt.Errorf("note is already deleted")
	}

	// Soft delete the note
	if err := s.noteRepo.Delete(noteID, userID); err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}

// RestoreNote restores a soft-deleted note
func (s *NoteService) RestoreNote(noteID, userID uuid.UUID) (*model.NoteResponse, error) {
	// Check if note exists and user has permission
	note, err := s.noteRepo.GetByIDAndUserID(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}
	if note == nil {
		return nil, fmt.Errorf("note not found")
	}

	// Check if note is actually deleted
	if !note.IsDeleted() {
		return nil, fmt.Errorf("note is not deleted")
	}

	// Restore the note
	if err := s.noteRepo.Restore(noteID, userID); err != nil {
		return nil, fmt.Errorf("failed to restore note: %w", err)
	}

	// Get updated note
	restoredNote, err := s.noteRepo.GetByIDAndUserID(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get restored note: %w", err)
	}

	return restoredNote.ToResponse(), nil
}

// HardDeleteNote permanently deletes a note
func (s *NoteService) HardDeleteNote(noteID, userID uuid.UUID) error {
	// Check if note exists and user has permission
	note, err := s.noteRepo.GetByIDAndUserID(noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}
	if note == nil {
		return fmt.Errorf("note not found")
	}

	// Hard delete the note
	if err := s.noteRepo.HardDelete(noteID, userID); err != nil {
		return fmt.Errorf("failed to hard delete note: %w", err)
	}

	return nil
}

// SearchNotes performs advanced search
func (s *NoteService) SearchNotes(userID uuid.UUID, req *model.NoteSearchRequest) (*model.NotesListResponse, error) {
	// Custom validation: ensure at least one search criteria is provided
	if req.Query == "" && req.IsPublic == nil && len(req.Tags) == 0 && req.DateFrom == nil && req.DateTo == nil {
		return nil, fmt.Errorf("validation error: at least one search criteria must be provided (query, is_public, tags, or date range)")
	}

	// Validate request structure
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Set defaults
	req.SetDefaults()

	// Perform search
	notes, total, err := s.noteRepo.Search(userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}

	// Convert search params to GetNotesParams for response
	params := &model.GetNotesParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Search:   req.Query,
		Status:   req.Status,
	}

	// Convert to response
	response := model.NewNotesListResponse(notes, total, params)
	return response, nil
}

// GetPublicNotes retrieves public notes (accessible to all users)
func (s *NoteService) GetPublicNotes(params *model.GetNotesParams) (*model.NotesListResponse, error) {
	// Validate parameters
	if err := s.validator.ValidateStruct(params); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Set defaults
	params.SetDefaults()

	// Get public notes
	notes, total, err := s.noteRepo.GetPublicNotes(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get public notes: %w", err)
	}

	// Convert to response
	response := model.NewNotesListResponse(notes, total, params)
	return response, nil
}

// BulkUpdateNotesStatus updates status for multiple notes
func (s *NoteService) BulkUpdateNotesStatus(userID uuid.UUID, req *model.BulkOperationRequest) error {
	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Validate operation
	if req.Operation != "update_status" {
		return fmt.Errorf("invalid operation: %s", req.Operation)
	}

	// Extract status from data
	statusData, ok := req.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid data format")
	}

	statusStr, ok := statusData["status"].(string)
	if !ok {
		return fmt.Errorf("status field is required")
	}

	status := model.NoteStatus(statusStr)
	if status != model.NoteStatusActive && status != model.NoteStatusDraft && status != model.NoteStatusDeleted {
		return fmt.Errorf("invalid status: %s", statusStr)
	}

	// Perform bulk update
	if err := s.noteRepo.BulkUpdateStatus(userID, req.NoteIDs, status); err != nil {
		return fmt.Errorf("failed to bulk update status: %w", err)
	}

	return nil
}

// GetNoteStats returns statistics for user's notes
func (s *NoteService) GetNoteStats(userID uuid.UUID) (map[string]interface{}, error) {
	// Get stats from repository
	stats, err := s.noteRepo.GetNoteStats(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note stats: %w", err)
	}

	return stats, nil
}

// ValidateNoteOwnership validates that a user owns a note
func (s *NoteService) ValidateNoteOwnership(noteID, userID uuid.UUID) error {
	note, err := s.noteRepo.GetByIDAndUserID(noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}
	if note == nil {
		return fmt.Errorf("note not found or access denied")
	}
	return nil
}

// DuplicateNote creates a copy of an existing note
func (s *NoteService) DuplicateNote(noteID, userID uuid.UUID) (*model.NoteResponse, error) {
	// Get original note
	originalNote, err := s.noteRepo.GetByIDAndUserID(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}
	if originalNote == nil {
		return nil, fmt.Errorf("note not found")
	}

	// Create duplicate
	duplicateNote := &model.Note{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     originalNote.Title + " (Copy)",
		Content:   originalNote.Content,
		Status:    model.NoteStatusDraft, // New copies start as draft
		Tags:      originalNote.Tags,
		IsPublic:  false, // Copies are private by default
		ViewCount: 0,
	}

	// Create in database
	if err := s.noteRepo.Create(duplicateNote); err != nil {
		return nil, fmt.Errorf("failed to create duplicate note: %w", err)
	}

	return duplicateNote.ToResponse(), nil
}

// validateNoteContent validates and sanitizes note content
func (s *NoteService) validateNoteContent(note *model.Note) error {
	// Trim whitespace
	note.Title = strings.TrimSpace(note.Title)

	// Validate title
	if note.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	if len(note.Title) > 255 {
		return fmt.Errorf("title too long (max 255 characters)")
	}

	// Validate content length
	if note.Content != nil && len(*note.Content) > 10000 {
		return fmt.Errorf("content too long (max 10000 characters)")
	}

	// Validate tags
	if note.Tags != nil {
		tags := note.GetTagsArray()
		if len(tags) > 10 {
			return fmt.Errorf("too many tags (max 10)")
		}

		for _, tag := range tags {
			if len(tag) > 50 {
				return fmt.Errorf("tag too long (max 50 characters): %s", tag)
			}
		}
	}

	return nil
}

// canUserAccessNote checks if a user can access a note
func (s *NoteService) canUserAccessNote(note *model.Note, userID uuid.UUID) bool {
	// Owner can always access
	if note.UserID == userID {
		return true
	}

	// Public notes can be accessed by anyone (if active)
	if note.IsPublic && note.Status == model.NoteStatusActive {
		return true
	}

	// Otherwise, no access
	return false
}

// GetNotesByTag retrieves notes by tag
func (s *NoteService) GetNotesByTag(userID uuid.UUID, tag string, params *model.GetNotesParams) (*model.NotesListResponse, error) {
	// Validate tag
	if tag == "" {
		return nil, fmt.Errorf("tag cannot be empty")
	}

	// Set tag in params
	params.Tags = tag

	// Get notes
	return s.GetUserNotes(userID, params)
}

// GetAllUserTags retrieves all unique tags for a user
func (s *NoteService) GetAllUserTags(userID uuid.UUID) ([]string, error) {
	// Get all active notes for user
	params := &model.GetNotesParams{
		Status:   "active",
		PageSize: 1000, // Large page size to get all notes
	}

	notes, _, err := s.noteRepo.GetByUserID(userID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get user notes: %w", err)
	}

	// Extract unique tags
	tagSet := make(map[string]bool)
	for _, note := range notes {
		tags := note.GetTagsArray()
		for _, tag := range tags {
			if tag != "" {
				tagSet[tag] = true
			}
		}
	}

	// Convert to slice
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags, nil
}

// ToggleNotePublicStatus toggles the public status of a note
func (s *NoteService) ToggleNotePublicStatus(noteID, userID uuid.UUID) (*model.NoteResponse, error) {
	// Get note
	note, err := s.noteRepo.GetByIDAndUserID(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}
	if note == nil {
		return nil, fmt.Errorf("note not found")
	}

	// Only active notes can be made public
	if note.Status != model.NoteStatusActive && !note.IsPublic {
		return nil, fmt.Errorf("only active notes can be made public")
	}

	// Toggle public status
	note.IsPublic = !note.IsPublic

	// Update in database
	if err := s.noteRepo.Update(note); err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return note.ToResponse(), nil
}
