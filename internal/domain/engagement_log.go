package domain

import (
	"time"

	"github.com/google/uuid"
)

// EngagementLog tracks inbound email interactions.
type EngagementLog struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	RawEmailText string    `json:"raw_email_text"`
	Status       string    `json:"status"`
	ReceivedAt   time.Time `json:"received_at"`
}

// Engagement log status constants.
const (
	EngagementStatusPending    = "pending"
	EngagementStatusProcessing = "processing"
	EngagementStatusCompleted  = "completed"
	EngagementStatusFailed     = "failed"
)
