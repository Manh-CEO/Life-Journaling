package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/life-journaling/core/internal/handler/dto"
	"github.com/life-journaling/core/internal/handler/middleware"
	"github.com/life-journaling/core/internal/usecase"
)

// MemoryHandler handles memory-related HTTP requests.
type MemoryHandler struct {
	memoryService *usecase.MemoryService
}

// NewMemoryHandler creates a new MemoryHandler.
func NewMemoryHandler(memoryService *usecase.MemoryService) *MemoryHandler {
	return &MemoryHandler{memoryService: memoryService}
}

// List returns paginated memories for the authenticated user.
func (h *MemoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.NewErrorResponse("unauthorized"))
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	params := usecase.PaginationParams{Limit: limit, Offset: offset}

	result, err := h.memoryService.List(r.Context(), userID, params)
	if err != nil {
		handleError(w, err)
		return
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}

	writeJSON(w, http.StatusOK, dto.NewPaginatedResponse(
		dto.ToMemoryResponses(result.Items),
		result.Total,
		params.Limit,
		params.Offset,
	))
}

// Create handles creation of a new memory.
func (h *MemoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.NewErrorResponse("unauthorized"))
		return
	}

	var req dto.CreateMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("invalid request body"))
		return
	}

	if req.Content == "" || req.EntryDate == "" {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("content and entry_date are required"))
		return
	}

	input := usecase.CreateMemoryInput{
		EntryDate: req.EntryDate,
		Location:  req.Location,
		Content:   req.Content,
		Sentiment: req.Sentiment,
	}

	memory, err := h.memoryService.Create(r.Context(), userID, input)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.NewSuccessResponse(dto.ToMemoryResponse(memory)))
}

// GetByID returns a single memory by ID.
func (h *MemoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.NewErrorResponse("unauthorized"))
		return
	}

	memoryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("invalid memory id"))
		return
	}

	memory, err := h.memoryService.GetByID(r.Context(), userID, memoryID)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.NewSuccessResponse(dto.ToMemoryResponse(memory)))
}

// Update handles updating an existing memory.
func (h *MemoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.NewErrorResponse("unauthorized"))
		return
	}

	memoryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("invalid memory id"))
		return
	}

	var req dto.UpdateMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("invalid request body"))
		return
	}

	input := usecase.UpdateMemoryInput{
		EntryDate: req.EntryDate,
		Location:  req.Location,
		Content:   req.Content,
		Sentiment: req.Sentiment,
	}

	memory, err := h.memoryService.Update(r.Context(), userID, memoryID, input)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.NewSuccessResponse(dto.ToMemoryResponse(memory)))
}

// Delete handles deletion of a memory.
func (h *MemoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.NewErrorResponse("unauthorized"))
		return
	}

	memoryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("invalid memory id"))
		return
	}

	if err := h.memoryService.Delete(r.Context(), userID, memoryID); err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.NewSuccessResponse(nil))
}
