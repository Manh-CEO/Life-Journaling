package domain

import (
	"time"

	"github.com/google/uuid"
)

// Portrait represents a yearly portrait photograph.
type Portrait struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	StoragePath    string    `json:"storage_path"`
	PortraitYear   int       `json:"portrait_year"`
	IsManualUpload bool      `json:"is_manual_upload"`
	CapturedAt     time.Time `json:"captured_at"`
}
