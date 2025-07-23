package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gonotes/internal/middleware"
	"gonotes/internal/model"
	"gonotes/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// NoteHandler handles HTTP requests for notes
type NoteHandler struct {
	noteService *service.NoteService
}

// NewNoteHandler creates a new note handler
func NewNoteHandler(noteService *service.NoteService) *NoteHandler {
	return &NoteHandler{
		noteService: noteService,
	}
}

// CreateNote handles POST /notes
func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Parse request body
	var req model.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid request body", nil, err.Error())
		return
	}

	// Create note
	note, err := h.noteService.CreateNote(userID, &req)
	if err != nil {
		if isValidationError(err) {
			sendResponse(w, http.StatusBadRequest, "error", "Validation error", nil, err.Error())
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to create note", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusCreated, "success", "Note created successfully", note, nil)
}

// GetNote handles GET /notes/{id}
func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	// Get note ID from URL
	noteIDStr := chi.URLParam(r, "id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid note ID", nil, err.Error())
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Get note
	note, err := h.noteService.GetNoteByID(noteID, userID)
	if err != nil {
		if err.Error() == "note not found" {
			sendResponse(w, http.StatusNotFound, "error", "Note not found", nil, nil)
			return
		}
		if err.Error() == "access denied" {
			sendResponse(w, http.StatusForbidden, "error", "Access denied", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to get note", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, "success", "Note retrieved successfully", note, nil)
}

// GetNotes handles GET /notes
func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Parse query parameters
	params := &model.GetNotesParams{
		Page:     getIntParam(r, "page", 1),
		PageSize: getIntParam(r, "page_size", 20),
		Search:   r.URL.Query().Get("search"),
		Status:   r.URL.Query().Get("status"),
		Tags:     r.URL.Query().Get("tags"),
		SortBy:   r.URL.Query().Get("sort_by"),
		SortDir:  r.URL.Query().Get("sort_dir"),
	}

	// Parse is_public parameter
	if isPublicStr := r.URL.Query().Get("is_public"); isPublicStr != "" {
		if isPublic, err := strconv.ParseBool(isPublicStr); err == nil {
			params.IsPublic = &isPublic
		}
	}

	// Get notes
	notes, err := h.noteService.GetUserNotes(userID, params)
	if err != nil {
		if isValidationError(err) {
			sendResponse(w, http.StatusBadRequest, "error", "Invalid parameters", nil, err.Error())
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to get notes", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, "success", "Notes retrieved successfully", notes, nil)
}

// UpdateNote handles PUT /notes/{id}
func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	// Get note ID from URL
	noteIDStr := chi.URLParam(r, "id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid note ID", nil, err.Error())
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Parse request body
	var req model.UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid request body", nil, err.Error())
		return
	}

	// Update note
	note, err := h.noteService.UpdateNote(noteID, userID, &req)
	if err != nil {
		if err.Error() == "note not found" {
			sendResponse(w, http.StatusNotFound, "error", "Note not found", nil, nil)
			return
		}
		if isValidationError(err) {
			sendResponse(w, http.StatusBadRequest, "error", "Validation error", nil, err.Error())
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to update note", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, "success", "Note updated successfully", note, nil)
}

// DeleteNote handles DELETE /notes/{id}
func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	// Get note ID from URL
	noteIDStr := chi.URLParam(r, "id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid note ID", nil, err.Error())
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Delete note
	if err := h.noteService.DeleteNote(noteID, userID); err != nil {
		if err.Error() == "note not found" {
			sendResponse(w, http.StatusNotFound, "error", "Note not found", nil, nil)
			return
		}
		if err.Error() == "note is already deleted" {
			sendResponse(w, http.StatusBadRequest, "error", "Note is already deleted", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to delete note", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, "success", "Note deleted successfully", nil, nil)
}

// RestoreNote handles POST /notes/{id}/restore
func (h *NoteHandler) RestoreNote(w http.ResponseWriter, r *http.Request) {
	// Get note ID from URL
	noteIDStr := chi.URLParam(r, "id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid note ID", nil, err.Error())
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Restore note
	note, err := h.noteService.RestoreNote(noteID, userID)
	if err != nil {
		if err.Error() == "note not found" {
			sendResponse(w, http.StatusNotFound, "error", "Note not found", nil, nil)
			return
		}
		if err.Error() == "note is not deleted" {
			sendResponse(w, http.StatusBadRequest, "error", "Note is not deleted", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to restore note", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, "success", "Note restored successfully", note, nil)
}

// HardDeleteNote handles DELETE /notes/{id}/hard
func (h *NoteHandler) HardDeleteNote(w http.ResponseWriter, r *http.Request) {
	// Get note ID from URL
	noteIDStr := chi.URLParam(r, "id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid note ID", nil, err.Error())
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Hard delete note
	if err := h.noteService.HardDeleteNote(noteID, userID); err != nil {
		if err.Error() == "note not found" {
			sendResponse(w, http.StatusNotFound, "error", "Note not found", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to permanently delete note", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, "success", "Note permanently deleted", nil, nil)
}

// SearchNotes handles POST /notes/search
func (h *NoteHandler) SearchNotes(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Parse request body
	var req model.NoteSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid request body", nil, err.Error())
		return
	}

	// Search notes
	notes, err := h.noteService.SearchNotes(userID, &req)
	if err != nil {
		if isValidationError(err) {
			sendResponse(w, http.StatusBadRequest, "error", "Validation error", nil, err.Error())
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to search notes", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, "success", "Search completed successfully", notes, nil)
}

// GetPublicNotes handles GET /notes/public
func (h *NoteHandler) GetPublicNotes(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	params := &model.GetNotesParams{
		Page:     getIntParam(r, "page", 1),
		PageSize: getIntParam(r, "page_size", 20),
		Search:   r.URL.Query().Get("search"),
		Tags:     r.URL.Query().Get("tags"),
		SortBy:   r.URL.Query().Get("sort_by"),
		SortDir:  r.URL.Query().Get("sort_dir"),
	}

	// Get public notes
	notes, err := h.noteService.GetPublicNotes(params)
	if err != nil {
		if isValidationError(err) {
			sendResponse(w, http.StatusBadRequest, "error", "Invalid parameters", nil, err.Error())
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to get public notes", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, "success", "Public notes retrieved successfully", notes, nil)
}

// BulkUpdateNotes handles POST /notes/bulk
func (h *NoteHandler) BulkUpdateNotes(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Parse request body
	var req model.BulkOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid request body", nil, err.Error())
		return
	}

	// Perform bulk operation
	if err := h.noteService.BulkUpdateNotesStatus(userID, &req); err != nil {
		if isValidationError(err) {
			sendResponse(w, http.StatusBadRequest, "error", "Validation error", nil, err.Error())
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to perform bulk operation", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, "success", "Bulk operation completed successfully", nil, nil)
}

// GetNoteStats handles GET /notes/stats
func (h *NoteHandler) GetNoteStats(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Get stats
	stats, err := h.noteService.GetNoteStats(userID)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to get note stats", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, "success", "Stats retrieved successfully", stats, nil)
}

// DuplicateNote handles POST /notes/{id}/duplicate
func (h *NoteHandler) DuplicateNote(w http.ResponseWriter, r *http.Request) {
	// Get note ID from URL
	noteIDStr := chi.URLParam(r, "id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid note ID", nil, err.Error())
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Duplicate note
	note, err := h.noteService.DuplicateNote(noteID, userID)
	if err != nil {
		if err.Error() == "note not found" {
			sendResponse(w, http.StatusNotFound, "error", "Note not found", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to duplicate note", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusCreated, "success", "Note duplicated successfully", note, nil)
}

// GetNotesByTag handles GET /notes/tag/{tag}
func (h *NoteHandler) GetNotesByTag(w http.ResponseWriter, r *http.Request) {
	// Get tag from URL
	tag := chi.URLParam(r, "tag")
	if tag == "" {
		sendResponse(w, http.StatusBadRequest, "error", "Tag parameter is required", nil, nil)
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Parse query parameters
	params := &model.GetNotesParams{
		Page:     getIntParam(r, "page", 1),
		PageSize: getIntParam(r, "page_size", 20),
		Search:   r.URL.Query().Get("search"),
		SortBy:   r.URL.Query().Get("sort_by"),
		SortDir:  r.URL.Query().Get("sort_dir"),
	}

	// Get notes by tag
	notes, err := h.noteService.GetNotesByTag(userID, tag, params)
	if err != nil {
		if isValidationError(err) {
			sendResponse(w, http.StatusBadRequest, "error", "Validation error", nil, err.Error())
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to get notes by tag", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, "success", "Notes retrieved successfully", notes, nil)
}

// GetUserTags handles GET /notes/tags
func (h *NoteHandler) GetUserTags(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Get all user tags
	tags, err := h.noteService.GetAllUserTags(userID)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to get user tags", nil, err.Error())
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, "success", "Tags retrieved successfully", map[string]interface{}{
		"tags": tags,
	}, nil)
}

// ToggleNotePublicStatus handles POST /notes/{id}/toggle-public
func (h *NoteHandler) ToggleNotePublicStatus(w http.ResponseWriter, r *http.Request) {
	// Get note ID from URL
	noteIDStr := chi.URLParam(r, "id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid note ID", nil, err.Error())
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "User not authenticated", nil, nil)
		return
	}

	// Toggle public status
	note, err := h.noteService.ToggleNotePublicStatus(noteID, userID)
	if err != nil {
		if err.Error() == "note not found" {
			sendResponse(w, http.StatusNotFound, "error", "Note not found", nil, nil)
			return
		}
		if err.Error() == "only active notes can be made public" {
			sendResponse(w, http.StatusBadRequest, "error", "Only active notes can be made public", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to toggle public status", nil, err.Error())
		return
	}

	// Send response
	message := "Note made private"
	if note.IsPublic {
		message = "Note made public"
	}
	sendResponse(w, http.StatusOK, "success", message, note, nil)
}

// Helper functions

// getIntParam extracts integer parameter from query string with default value
func getIntParam(r *http.Request, key string, defaultValue int) int {
	valueStr := r.URL.Query().Get(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// isValidationError checks if error is a validation error
func isValidationError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return contains(errMsg, "validation error") ||
		contains(errMsg, "invalid") ||
		contains(errMsg, "required") ||
		contains(errMsg, "too long") ||
		contains(errMsg, "too short") ||
		contains(errMsg, "cannot be empty")
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
