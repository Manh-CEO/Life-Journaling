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

// PortraitHandler handles portrait-related HTTP requests.
type PortraitHandler struct {
	portraitService *usecase.PortraitService
}

// NewPortraitHandler creates a new PortraitHandler.
func NewPortraitHandler(portraitService *usecase.PortraitService) *PortraitHandler {
	return &PortraitHandler{portraitService: portraitService}
}

// List returns paginated portraits for the authenticated user.
func (h *PortraitHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.NewErrorResponse("unauthorized"))
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	params := usecase.PaginationParams{Limit: limit, Offset: offset}

	result, err := h.portraitService.List(r.Context(), userID, params)
	if err != nil {
		handleError(w, err)
		return
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}

	writeJSON(w, http.StatusOK, dto.NewPaginatedResponse(
		dto.ToPortraitResponses(result.Items),
		result.Total,
		params.Limit,
		params.Offset,
	))
}

// Create handles creation of a new portrait.
func (h *PortraitHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.NewErrorResponse("unauthorized"))
		return
	}

	var req dto.CreatePortraitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("invalid request body"))
		return
	}

	if req.StoragePath == "" || req.PortraitYear == 0 {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("storage_path and portrait_year are required"))
		return
	}

	input := usecase.CreatePortraitInput{
		StoragePath:    req.StoragePath,
		PortraitYear:   req.PortraitYear,
		IsManualUpload: req.IsManualUpload,
		CapturedAt:     req.CapturedAt,
	}

	portrait, err := h.portraitService.Create(r.Context(), userID, input)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.NewSuccessResponse(dto.ToPortraitResponse(portrait)))
}

// GetLatest returns the latest portrait for the authenticated user.
func (h *PortraitHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.NewErrorResponse("unauthorized"))
		return
	}

	portrait, err := h.portraitService.GetLatest(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.NewSuccessResponse(dto.ToPortraitResponse(portrait)))
}

// Delete handles deletion of a portrait.
func (h *PortraitHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.NewErrorResponse("unauthorized"))
		return
	}

	portraitID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("invalid portrait id"))
		return
	}

	if err := h.portraitService.Delete(r.Context(), userID, portraitID); err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.NewSuccessResponse(nil))
}
