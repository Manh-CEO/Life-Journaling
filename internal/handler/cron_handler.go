package handler

import (
	"log/slog"
	"net/http"

	"github.com/life-journaling/core/internal/handler/dto"
	"github.com/life-journaling/core/internal/usecase"
)

// CronHandler handles cron trigger HTTP requests.
type CronHandler struct {
	engagementService *usecase.EngagementService
}

// NewCronHandler creates a new CronHandler.
func NewCronHandler(engagementService *usecase.EngagementService) *CronHandler {
	return &CronHandler{engagementService: engagementService}
}

// Hourly handles the hourly cron trigger from QStash.
func (h *CronHandler) Hourly(w http.ResponseWriter, r *http.Request) {
	if err := h.engagementService.SendHourlyPrompts(r.Context()); err != nil {
		slog.Error("hourly cron failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, dto.NewErrorResponse("failed to send hourly prompts"))
		return
	}

	writeJSON(w, http.StatusOK, dto.NewSuccessResponse(map[string]string{
		"status": "hourly prompts sent",
	}))
}

// Annual handles the annual anchor date cron trigger from QStash.
func (h *CronHandler) Annual(w http.ResponseWriter, r *http.Request) {
	if err := h.engagementService.SendAnchorDateEmails(r.Context()); err != nil {
		slog.Error("annual cron failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, dto.NewErrorResponse("failed to send anchor date emails"))
		return
	}

	writeJSON(w, http.StatusOK, dto.NewSuccessResponse(map[string]string{
		"status": "anchor date emails sent",
	}))
}
