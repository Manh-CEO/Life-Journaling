package domain

import (
	"time"

	"github.com/google/uuid"
)

// Memory represents a single journal entry or memory.
type Memory struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	EntryDate     time.Time `json:"entry_date"`
	Location      string    `json:"location"`
	Content       string    `json:"content"`
	Sentiment     string    `json:"sentiment"`
	IsManualEntry bool      `json:"is_manual_entry"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
