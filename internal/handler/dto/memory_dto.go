package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/life-journaling/core/internal/domain"
)

// MemoryResponse represents a memory in API responses.
type MemoryResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	EntryDate     string    `json:"entry_date"`
	Location      string    `json:"location"`
	Content       string    `json:"content"`
	Sentiment     string    `json:"sentiment"`
	IsManualEntry bool      `json:"is_manual_entry"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateMemoryRequest represents a request to create a memory.
type CreateMemoryRequest struct {
	EntryDate string `json:"entry_date" validate:"required"`
	Location  string `json:"location"`
	Content   string `json:"content" validate:"required,min=1"`
	Sentiment string `json:"sentiment" validate:"omitempty,oneof=positive negative neutral mixed"`
}

// UpdateMemoryRequest represents a request to update a memory.
type UpdateMemoryRequest struct {
	EntryDate *string `json:"entry_date"`
	Location  *string `json:"location"`
	Content   *string `json:"content" validate:"omitempty,min=1"`
	Sentiment *string `json:"sentiment" validate:"omitempty,oneof=positive negative neutral mixed"`
}

// ToMemoryResponse converts a domain Memory to a MemoryResponse DTO.
func ToMemoryResponse(memory domain.Memory) MemoryResponse {
	return MemoryResponse{
		ID:            memory.ID,
		UserID:        memory.UserID,
		EntryDate:     memory.EntryDate.Format("2006-01-02"),
		Location:      memory.Location,
		Content:       memory.Content,
		Sentiment:     memory.Sentiment,
		IsManualEntry: memory.IsManualEntry,
		CreatedAt:     memory.CreatedAt,
		UpdatedAt:     memory.UpdatedAt,
	}
}

// ToMemoryResponses converts a slice of domain Memories to MemoryResponse DTOs.
func ToMemoryResponses(memories []domain.Memory) []MemoryResponse {
	responses := make([]MemoryResponse, 0, len(memories))
	for _, m := range memories {
		responses = append(responses, ToMemoryResponse(m))
	}
	return responses
}
