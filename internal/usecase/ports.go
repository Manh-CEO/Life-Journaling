package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/life-journaling/core/internal/domain"
)

// PaginationParams holds pagination parameters.
type PaginationParams struct {
	Limit  int
	Offset int
}

// PaginatedResult holds paginated query results.
type PaginatedResult[T any] struct {
	Items []T
	Total int
}

// IUserRepository defines data access operations for users.
type IUserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Create(ctx context.Context, user domain.User) (domain.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, update UserProfileUpdate) (domain.User, error)
	GetUsersForPrompt(ctx context.Context, dayOfWeek int, hour int) ([]domain.User, error)
	GetUsersForAnchorDate(ctx context.Context, month int, day int) ([]domain.User, error)
}

// UserProfileUpdate contains updatable user fields.
type UserProfileUpdate struct {
	Timezone        *string
	AnchorDate      *string
	PromptDayOfWeek *int
	PromptHour      *int
}

// IMemoryRepository defines data access operations for memories.
type IMemoryRepository interface {
	Create(ctx context.Context, memory domain.Memory) (domain.Memory, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Memory, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, params PaginationParams) (PaginatedResult[domain.Memory], error)
	Update(ctx context.Context, memory domain.Memory) (domain.Memory, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// IEngagementLogRepository defines data access operations for engagement logs.
type IEngagementLogRepository interface {
	Create(ctx context.Context, log domain.EngagementLog) (domain.EngagementLog, error)
	GetPendingByUserID(ctx context.Context, userID uuid.UUID) ([]domain.EngagementLog, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// IPortraitRepository defines data access operations for portraits.
type IPortraitRepository interface {
	Create(ctx context.Context, portrait domain.Portrait) (domain.Portrait, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, params PaginationParams) (PaginatedResult[domain.Portrait], error)
	GetLatestByUserID(ctx context.Context, userID uuid.UUID) (domain.Portrait, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// IEmailProvider defines email sending capabilities.
type IEmailProvider interface {
	SendPrompt(ctx context.Context, toEmail string, subject string, body string) error
}

// ILLMProvider defines LLM processing capabilities.
type ILLMProvider interface {
	ExtractMemoryData(ctx context.Context, rawText string) (domain.Memory, error)
}
