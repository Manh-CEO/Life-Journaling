package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/life-journaling/core/internal/domain"
)

// PortraitResponse represents a portrait in API responses.
type PortraitResponse struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	StoragePath    string    `json:"storage_path"`
	PortraitYear   int       `json:"portrait_year"`
	IsManualUpload bool      `json:"is_manual_upload"`
	CapturedAt     time.Time `json:"captured_at"`
}

// CreatePortraitRequest represents a request to create a portrait.
type CreatePortraitRequest struct {
	StoragePath    string `json:"storage_path" validate:"required"`
	PortraitYear   int    `json:"portrait_year" validate:"required,min=1900,max=2100"`
	IsManualUpload bool   `json:"is_manual_upload"`
	CapturedAt     string `json:"captured_at"`
}

// ToPortraitResponse converts a domain Portrait to a PortraitResponse DTO.
func ToPortraitResponse(portrait domain.Portrait) PortraitResponse {
	return PortraitResponse{
		ID:             portrait.ID,
		UserID:         portrait.UserID,
		StoragePath:    portrait.StoragePath,
		PortraitYear:   portrait.PortraitYear,
		IsManualUpload: portrait.IsManualUpload,
		CapturedAt:     portrait.CapturedAt,
	}
}

// ToPortraitResponses converts a slice of domain Portraits to PortraitResponse DTOs.
func ToPortraitResponses(portraits []domain.Portrait) []PortraitResponse {
	responses := make([]PortraitResponse, 0, len(portraits))
	for _, p := range portraits {
		responses = append(responses, ToPortraitResponse(p))
	}
	return responses
}
