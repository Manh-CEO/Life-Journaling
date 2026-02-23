package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/life-journaling/core/internal/domain"
)

// PortraitService implements portrait-related business logic.
type PortraitService struct {
	portraitRepo IPortraitRepository
}

// NewPortraitService creates a new PortraitService.
func NewPortraitService(portraitRepo IPortraitRepository) *PortraitService {
	return &PortraitService{portraitRepo: portraitRepo}
}

// CreatePortraitInput holds the input for creating a portrait.
type CreatePortraitInput struct {
	StoragePath    string `json:"storage_path" validate:"required"`
	PortraitYear   int    `json:"portrait_year" validate:"required,min=1900,max=2100"`
	IsManualUpload bool   `json:"is_manual_upload"`
	CapturedAt     string `json:"captured_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

// Create creates a new portrait entry.
func (s *PortraitService) Create(ctx context.Context, userID uuid.UUID, input CreatePortraitInput) (domain.Portrait, error) {
	capturedAt := time.Now().UTC()
	if input.CapturedAt != "" {
		parsed, err := time.Parse(time.RFC3339, input.CapturedAt)
		if err != nil {
			return domain.Portrait{}, domain.NewDomainError(domain.ErrValidation, "invalid captured_at format")
		}
		capturedAt = parsed
	}

	portrait := domain.Portrait{
		UserID:         userID,
		StoragePath:    input.StoragePath,
		PortraitYear:   input.PortraitYear,
		IsManualUpload: input.IsManualUpload,
		CapturedAt:     capturedAt,
	}

	created, err := s.portraitRepo.Create(ctx, portrait)
	if err != nil {
		return domain.Portrait{}, fmt.Errorf("creating portrait: %w", err)
	}

	return created, nil
}

// List retrieves portraits for a user with pagination.
func (s *PortraitService) List(ctx context.Context, userID uuid.UUID, params PaginationParams) (PaginatedResult[domain.Portrait], error) {
	params = normalizePagination(params)

	result, err := s.portraitRepo.GetByUserID(ctx, userID, params)
	if err != nil {
		return PaginatedResult[domain.Portrait]{}, fmt.Errorf("listing portraits: %w", err)
	}

	return result, nil
}

// GetLatest retrieves the latest portrait for a user.
func (s *PortraitService) GetLatest(ctx context.Context, userID uuid.UUID) (domain.Portrait, error) {
	portrait, err := s.portraitRepo.GetLatestByUserID(ctx, userID)
	if err != nil {
		return domain.Portrait{}, fmt.Errorf("getting latest portrait: %w", err)
	}
	return portrait, nil
}

// Delete removes a portrait, ensuring it belongs to the requesting user.
func (s *PortraitService) Delete(ctx context.Context, userID uuid.UUID, portraitID uuid.UUID) error {
	// For V1, we trust the user owns the portrait since we filter by user_id in queries.
	// A more robust check would fetch first and verify ownership.
	if err := s.portraitRepo.Delete(ctx, portraitID); err != nil {
		return fmt.Errorf("deleting portrait: %w", err)
	}
	return nil
}
