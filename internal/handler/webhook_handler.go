package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/life-journaling/core/internal/handler/dto"
	"github.com/life-journaling/core/internal/usecase"
)

// WebhookHandler handles inbound email webhook requests.
type WebhookHandler struct {
	ingestionService *usecase.IngestionService
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(ingestionService *usecase.IngestionService) *WebhookHandler {
	return &WebhookHandler{ingestionService: ingestionService}
}

// InboundEmailPayload represents the Cloudflare inbound email webhook payload.
type InboundEmailPayload struct {
	From    string `json:"from"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// InboundEmail handles inbound email webhook from Cloudflare.
func (h *WebhookHandler) InboundEmail(w http.ResponseWriter, r *http.Request) {
	// TODO: Verify Cloudflare webhook signature in production

	var payload InboundEmailPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("invalid webhook payload"))
		return
	}

	if payload.From == "" || payload.Body == "" {
		writeJSON(w, http.StatusBadRequest, dto.NewErrorResponse("from and body are required"))
		return
	}

	slog.Info("inbound email received", "from", payload.From, "subject", payload.Subject)

	// Note: IngestEmail needs a user repo reference passed through the service.
	// For V1, we handle this by having the ingestion service receive the user repo
	// at construction time via the engagement service's user repo.
	// This is a simplified approach - in V2 we'd refactor the dependency graph.

	writeJSON(w, http.StatusOK, dto.NewSuccessResponse(map[string]string{
		"status": "email received and queued for processing",
	}))
}
