package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/life-journaling/core/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of IUserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, update UserProfileUpdate) (domain.User, error) {
	args := m.Called(ctx, id, update)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepository) GetUsersForPrompt(ctx context.Context, dayOfWeek int, hour int) ([]domain.User, error) {
	args := m.Called(ctx, dayOfWeek, hour)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserRepository) GetUsersForAnchorDate(ctx context.Context, month int, day int) ([]domain.User, error) {
	args := m.Called(ctx, month, day)
	return args.Get(0).([]domain.User), args.Error(1)
}

func TestUserService_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		userID   uuid.UUID
		mockUser domain.User
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "success",
			userID: uuid.New(),
			mockUser: domain.User{
				ID:    uuid.New(),
				Email: "test@example.com",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:     "user not found",
			userID:   uuid.New(),
			mockUser: domain.User{},
			mockErr:  domain.ErrNotFound,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			service := NewUserService(mockRepo)

			mockRepo.On("GetByID", mock.Anything, tt.userID).Return(tt.mockUser, tt.mockErr)

			user, err := service.GetByID(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, domain.User{}, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockUser, user)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetOrCreateByEmail(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	existingUser := domain.User{
		ID:    userID,
		Email: email,
	}

	t.Run("user exists", func(t *testing.T) {
		mockRepo := &MockUserRepository{}
		service := NewUserService(mockRepo)

		mockRepo.On("GetByID", mock.Anything, userID).Return(existingUser, nil)

		user, err := service.GetOrCreateByEmail(context.Background(), userID, email)

		assert.NoError(t, err)
		assert.Equal(t, existingUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user does not exist, create new", func(t *testing.T) {
		mockRepo := &MockUserRepository{}
		service := NewUserService(mockRepo)

		mockRepo.On("GetByID", mock.Anything, userID).Return(domain.User{}, domain.ErrNotFound)
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(user domain.User) bool {
			return user.ID == userID && user.Email == email
		})).Return(existingUser, nil)

		user, err := service.GetOrCreateByEmail(context.Background(), userID, email)

		assert.NoError(t, err)
		assert.Equal(t, existingUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create fails", func(t *testing.T) {
		mockRepo := &MockUserRepository{}
		service := NewUserService(mockRepo)

		mockRepo.On("GetByID", mock.Anything, userID).Return(domain.User{}, domain.ErrNotFound)
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(domain.User{}, errors.New("create failed"))

		_, err := service.GetOrCreateByEmail(context.Background(), userID, email)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateProfile(t *testing.T) {
	userID := uuid.New()
	updatedUser := domain.User{
		ID:              userID,
		Email:           "test@example.com",
		Timezone:        "America/New_York",
		PromptDayOfWeek: 1,
		PromptHour:      8,
	}

	tests := []struct {
		name     string
		update   UserProfileUpdate
		mockErr  error
		wantErr  bool
		errType  error
	}{
		{
			name: "success",
			update: UserProfileUpdate{
				Timezone:        &[]string{"America/New_York"}[0],
				PromptDayOfWeek: &[]int{1}[0],
				PromptHour:      &[]int{8}[0],
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "invalid prompt_day_of_week",
			update: UserProfileUpdate{
				PromptDayOfWeek: &[]int{7}[0], // Invalid: should be 0-6
			},
			wantErr: true,
			errType: domain.ErrValidation,
		},
		{
			name: "invalid prompt_hour",
			update: UserProfileUpdate{
				PromptHour: &[]int{25}[0], // Invalid: should be 0-23
			},
			wantErr: true,
			errType: domain.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			service := NewUserService(mockRepo)

			if tt.wantErr && tt.errType == domain.ErrValidation {
				// Validation error before calling repo
				mockRepo.On("GetByID", mock.Anything, userID).Return(updatedUser, nil)
			} else {
				mockRepo.On("GetByID", mock.Anything, userID).Return(updatedUser, nil)
				mockRepo.On("UpdateProfile", mock.Anything, userID, tt.update).Return(updatedUser, tt.mockErr)
			}

			user, err := service.UpdateProfile(context.Background(), userID, tt.update)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.True(t, errors.Is(err, tt.errType))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, updatedUser, user)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
