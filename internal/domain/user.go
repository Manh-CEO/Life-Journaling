package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user of the journaling platform.
type User struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	Timezone       string    `json:"timezone"`
	AnchorDate     *time.Time `json:"anchor_date,omitempty"`
	PromptDayOfWeek int      `json:"prompt_day_of_week"`
	PromptHour     int       `json:"prompt_hour"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
