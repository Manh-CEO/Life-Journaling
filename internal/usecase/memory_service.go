package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/life-journaling/core/internal/domain"
)

// MemoryService implements memory-related business logic.
type MemoryService struct {
	memoryRepo IMemoryRepository
}

// NewMemoryService creates a new MemoryService.
func NewMemoryService(memoryRepo IMemoryRepository) *MemoryService {
	return &MemoryService{memoryRepo: memoryRepo}
}

// CreateMemoryInput holds the input for creating a memory.
type CreateMemoryInput struct {
	EntryDate string `json:"entry_date" validate:"required,datetime=2006-01-02"`
	Location  string `json:"location"`
	Content   string `json:"content" validate:"required,min=1"`
	Sentiment string `json:"sentiment" validate:"omitempty,oneof=positive negative neutral mixed"`
}

// Create creates a new manual memory entry.
func (s *MemoryService) Create(ctx context.Context, userID uuid.UUID, input CreateMemoryInput) (domain.Memory, error) {
	// Validate input
	if input.Content == "" {
		return domain.Memory{}, domain.NewDomainError(domain.ErrValidation, "content is required")
	}

	entryDate, err := time.Parse("2006-01-02", input.EntryDate)
	if err != nil {
		return domain.Memory{}, domain.NewDomainError(domain.ErrValidation, "invalid entry_date format, expected YYYY-MM-DD")
	}

	sentiment := input.Sentiment
	if sentiment == "" {
		sentiment = "neutral"
	}

	memory := domain.Memory{
		UserID:        userID,
		EntryDate:     entryDate,
		Location:      input.Location,
		Content:       input.Content,
		Sentiment:     sentiment,
		IsManualEntry: true,
	}

	created, err := s.memoryRepo.Create(ctx, memory)
	if err != nil {
		return domain.Memory{}, fmt.Errorf("creating memory: %w", err)
	}

	return created, nil
}

// GetByID retrieves a memory by ID, ensuring it belongs to the requesting user.
func (s *MemoryService) GetByID(ctx context.Context, userID uuid.UUID, memoryID uuid.UUID) (domain.Memory, error) {
	memory, err := s.memoryRepo.GetByID(ctx, memoryID)
	if err != nil {
		return domain.Memory{}, fmt.Errorf("getting memory: %w", err)
	}

	if memory.UserID != userID {
		return domain.Memory{}, domain.ErrForbidden
	}

	return memory, nil
}

// List retrieves memories for a user with pagination.
func (s *MemoryService) List(ctx context.Context, userID uuid.UUID, params PaginationParams) (PaginatedResult[domain.Memory], error) {
	params = normalizePagination(params)

	result, err := s.memoryRepo.GetByUserID(ctx, userID, params)
	if err != nil {
		return PaginatedResult[domain.Memory]{}, fmt.Errorf("listing memories: %w", err)
	}

	return result, nil
}

// UpdateMemoryInput holds the input for updating a memory.
type UpdateMemoryInput struct {
	EntryDate *string `json:"entry_date" validate:"omitempty,datetime=2006-01-02"`
	Location  *string `json:"location"`
	Content   *string `json:"content" validate:"omitempty,min=1"`
	Sentiment *string `json:"sentiment" validate:"omitempty,oneof=positive negative neutral mixed"`
}

// Update updates an existing memory.
func (s *MemoryService) Update(ctx context.Context, userID uuid.UUID, memoryID uuid.UUID, input UpdateMemoryInput) (domain.Memory, error) {
	existing, err := s.memoryRepo.GetByID(ctx, memoryID)
	if err != nil {
		return domain.Memory{}, fmt.Errorf("getting memory for update: %w", err)
	}

	if existing.UserID != userID {
		return domain.Memory{}, domain.ErrForbidden
	}

	// Build updated memory (immutable: new object)
	updated := domain.Memory{
		ID:            existing.ID,
		UserID:        existing.UserID,
		EntryDate:     existing.EntryDate,
		Location:      existing.Location,
		Content:       existing.Content,
		Sentiment:     existing.Sentiment,
		IsManualEntry: existing.IsManualEntry,
		CreatedAt:     existing.CreatedAt,
	}

	if input.EntryDate != nil {
		parsed, err := time.Parse("2006-01-02", *input.EntryDate)
		if err != nil {
			return domain.Memory{}, domain.NewDomainError(domain.ErrValidation, "invalid entry_date format")
		}
		updated.EntryDate = parsed
	}
	if input.Location != nil {
		updated.Location = *input.Location
	}
	if input.Content != nil {
		updated.Content = *input.Content
	}
	if input.Sentiment != nil {
		updated.Sentiment = *input.Sentiment
	}

	result, err := s.memoryRepo.Update(ctx, updated)
	if err != nil {
		return domain.Memory{}, fmt.Errorf("updating memory: %w", err)
	}

	return result, nil
}

// Delete removes a memory, ensuring it belongs to the requesting user.
func (s *MemoryService) Delete(ctx context.Context, userID uuid.UUID, memoryID uuid.UUID) error {
	memory, err := s.memoryRepo.GetByID(ctx, memoryID)
	if err != nil {
		return fmt.Errorf("getting memory for delete: %w", err)
	}

	if memory.UserID != userID {
		return domain.ErrForbidden
	}

	if err := s.memoryRepo.Delete(ctx, memoryID); err != nil {
		return fmt.Errorf("deleting memory: %w", err)
	}

	return nil
}

// normalizePagination applies defaults and limits to pagination parameters.
func normalizePagination(params PaginationParams) PaginationParams {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	return params
}
