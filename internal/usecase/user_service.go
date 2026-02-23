package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/life-journaling/core/internal/domain"
)

// UserService implements user-related business logic.
type UserService struct {
	userRepo IUserRepository
}

// NewUserService creates a new UserService.
func NewUserService(userRepo IUserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetByID retrieves a user by their ID.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("getting user by id: %w", err)
	}
	return user, nil
}

// GetOrCreateByEmail retrieves a user by email or creates a new one.
func (s *UserService) GetOrCreateByEmail(ctx context.Context, id uuid.UUID, email string) (domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err == nil {
		return user, nil
	}

	// User does not exist, create a new one
	newUser := domain.User{
		ID:              id,
		Email:           email,
		Timezone:        "UTC",
		PromptDayOfWeek: 0,
		PromptHour:      9,
	}

	created, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		return domain.User{}, fmt.Errorf("creating user: %w", err)
	}

	return created, nil
}

// UpdateProfile updates a user's profile settings.
func (s *UserService) UpdateProfile(ctx context.Context, id uuid.UUID, update UserProfileUpdate) (domain.User, error) {
	// Verify user exists
	if _, err := s.userRepo.GetByID(ctx, id); err != nil {
		return domain.User{}, fmt.Errorf("user not found: %w", err)
	}

	// Validate prompt_day_of_week
	if update.PromptDayOfWeek != nil {
		dow := *update.PromptDayOfWeek
		if dow < 0 || dow > 6 {
			return domain.User{}, domain.NewDomainError(
				domain.ErrValidation,
				"prompt_day_of_week must be between 0 and 6",
			)
		}
	}

	// Validate prompt_hour
	if update.PromptHour != nil {
		hour := *update.PromptHour
		if hour < 0 || hour > 23 {
			return domain.User{}, domain.NewDomainError(
				domain.ErrValidation,
				"prompt_hour must be between 0 and 23",
			)
		}
	}

	updated, err := s.userRepo.UpdateProfile(ctx, id, update)
	if err != nil {
		return domain.User{}, fmt.Errorf("updating user profile: %w", err)
	}

	return updated, nil
}
