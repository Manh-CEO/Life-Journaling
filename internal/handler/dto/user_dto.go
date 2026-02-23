package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/life-journaling/core/internal/domain"
)

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	Timezone        string     `json:"timezone"`
	AnchorDate      *time.Time `json:"anchor_date,omitempty"`
	PromptDayOfWeek int        `json:"prompt_day_of_week"`
	PromptHour      int        `json:"prompt_hour"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// UpdateUserRequest represents a request to update user profile.
type UpdateUserRequest struct {
	Timezone        *string `json:"timezone" validate:"omitempty,timezone"`
	AnchorDate      *string `json:"anchor_date" validate:"omitempty,datetime=2006-01-02"`
	PromptDayOfWeek *int    `json:"prompt_day_of_week" validate:"omitempty,min=0,max=6"`
	PromptHour      *int    `json:"prompt_hour" validate:"omitempty,min=0,max=23"`
}

// ToUserResponse converts a domain User to a UserResponse DTO.
func ToUserResponse(user domain.User) UserResponse {
	return UserResponse{
		ID:              user.ID,
		Email:           user.Email,
		Timezone:        user.Timezone,
		AnchorDate:      user.AnchorDate,
		PromptDayOfWeek: user.PromptDayOfWeek,
		PromptHour:      user.PromptHour,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}
