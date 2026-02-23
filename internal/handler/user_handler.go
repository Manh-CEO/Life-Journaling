package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/life-journaling/core/internal/domain"
	"github.com/life-journaling/core/internal/handler/dto"
	"github.com/life-journaling/core/internal/handler/middleware"
	"github.com/life-journaling/core/internal/usecase"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService *usecase.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *usecase.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetMe returns the current authenticated user's profile.
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.NewErrorResponse("unauthorized"))
		return
	}

	email, _ := middleware.GetUserEmail(r.Context())

	user, err := h.userService.GetOrCreateByEmail(r.Context(), userID, email)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.NewSuccessResponse(dto.ToUserResponse(user)))
}

// UpdateMe updates the current authenticated user's profile.
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.NewErrorResponse("unauthorized"))
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("invalid request body"))
		return
	}

	update := usecase.UserProfileUpdate{
		Timezone:        req.Timezone,
		AnchorDate:      req.AnchorDate,
		PromptDayOfWeek: req.PromptDayOfWeek,
		PromptHour:      req.PromptHour,
	}

	user, err := h.userService.UpdateProfile(r.Context(), userID, update)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.NewSuccessResponse(dto.ToUserResponse(user)))
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

// handleError maps domain errors to HTTP status codes and writes an error response.
func handleError(w http.ResponseWriter, err error) {
	slog.Error("handler error", "error", err)

	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		switch {
		case errors.Is(domainErr.Err, domain.ErrNotFound):
			writeJSON(w, http.StatusNotFound, dto.NewErrorResponse(domainErr.Message))
		case errors.Is(domainErr.Err, domain.ErrValidation):
			writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse(domainErr.Message))
		case errors.Is(domainErr.Err, domain.ErrUnauthorized):
			writeJSON(w, http.StatusUnauthorized, dto.NewErrorResponse(domainErr.Message))
		case errors.Is(domainErr.Err, domain.ErrForbidden):
			writeJSON(w, http.StatusForbidden, dto.NewErrorResponse(domainErr.Message))
		case errors.Is(domainErr.Err, domain.ErrAlreadyExists):
			writeJSON(w, http.StatusConflict, dto.NewErrorResponse(domainErr.Message))
		default:
			writeJSON(w, http.StatusInternalServerError, dto.NewErrorResponse("internal server error"))
		}
		return
	}

	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, dto.NewErrorResponse("not found"))
		return
	}
	if errors.Is(err, domain.ErrValidation) {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("validation failed"))
		return
	}

	writeJSON(w, http.StatusInternalServerError, dto.NewErrorResponse("internal server error"))
}
