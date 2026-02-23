package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/life-journaling/core/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPortraitRepository is a mock implementation of IPortraitRepository
type MockPortraitRepository struct {
	mock.Mock
}

func (m *MockPortraitRepository) Create(ctx context.Context, portrait domain.Portrait) (domain.Portrait, error) {
	args := m.Called(ctx, portrait)
	return args.Get(0).(domain.Portrait), args.Error(1)
}

func (m *MockPortraitRepository) GetByUserID(ctx context.Context, userID uuid.UUID, params PaginationParams) (PaginatedResult[domain.Portrait], error) {
	args := m.Called(ctx, userID, params)
	return args.Get(0).(PaginatedResult[domain.Portrait]), args.Error(1)
}

func (m *MockPortraitRepository) GetLatestByUserID(ctx context.Context, userID uuid.UUID) (domain.Portrait, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(domain.Portrait), args.Error(1)
}

func (m *MockPortraitRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestPortraitService_Create(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name     string
		input    CreatePortraitInput
		mockErr  error
		wantErr  bool
		errType  error
	}{
		{
			name: "success with all fields",
			input: CreatePortraitInput{
				StoragePath:    "portraits/2024/user123/portrait.jpg",
				PortraitYear:   2024,
				IsManualUpload: true,
				CapturedAt:     "2024-01-15T10:30:00Z",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "success with minimal fields",
			input: CreatePortraitInput{
				StoragePath:  "portraits/2024/user123/portrait.jpg",
				PortraitYear: 2024,
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "success with current timestamp",
			input: CreatePortraitInput{
				StoragePath:  "portraits/2024/user123/portrait.jpg",
				PortraitYear: 2024,
				CapturedAt:   "",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "invalid captured_at format",
			input: CreatePortraitInput{
				StoragePath:  "portraits/2024/user123/portrait.jpg",
				PortraitYear: 2024,
				CapturedAt:   "invalid-date",
			},
			wantErr: true,
			errType: domain.ErrValidation,
		},
		{
			name: "repo create fails",
			input: CreatePortraitInput{
				StoragePath:  "portraits/2024/user123/portrait.jpg",
				PortraitYear: 2024,
			},
			mockErr: errors.New("repo error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockPortraitRepository{}
			service := NewPortraitService(mockRepo)

			if !tt.wantErr || tt.errType != domain.ErrValidation {
				expectedPortrait := domain.Portrait{
					UserID:         userID,
					StoragePath:    tt.input.StoragePath,
					PortraitYear:   tt.input.PortraitYear,
					IsManualUpload: tt.input.IsManualUpload,
				}

				if tt.input.CapturedAt != "" && !tt.wantErr {
					parsed, _ := time.Parse(time.RFC3339, tt.input.CapturedAt)
					expectedPortrait.CapturedAt = parsed
				}

				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p domain.Portrait) bool {
					return p.UserID == userID && p.StoragePath == tt.input.StoragePath
				})).Return(expectedPortrait, tt.mockErr)
			}

			portrait, err := service.Create(context.Background(), userID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.True(t, errors.Is(err, tt.errType))
				}
				assert.Equal(t, domain.Portrait{}, portrait)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, userID, portrait.UserID)
				assert.Equal(t, tt.input.StoragePath, portrait.StoragePath)
				assert.Equal(t, tt.input.PortraitYear, portrait.PortraitYear)
				assert.Equal(t, tt.input.IsManualUpload, portrait.IsManualUpload)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPortraitService_List(t *testing.T) {
	userID := uuid.New()
	params := PaginationParams{Limit: 10, Offset: 0}

	expectedResult := PaginatedResult[domain.Portrait]{
		Items: []domain.Portrait{
			{
				ID:             uuid.New(),
				UserID:         userID,
				StoragePath:    "path1.jpg",
				PortraitYear:   2024,
				IsManualUpload: true,
			},
			{
				ID:             uuid.New(),
				UserID:         userID,
				StoragePath:    "path2.jpg",
				PortraitYear:   2023,
				IsManualUpload: false,
			},
		},
		Total: 2,
	}

	tests := []struct {
		name     string
		params   PaginationParams
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "success",
			params:  params,
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "repo error",
			params:  params,
			mockErr: errors.New("repo error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockPortraitRepository{}
			service := NewPortraitService(mockRepo)

			normalizedParams := normalizePagination(tt.params)
			mockRepo.On("GetByUserID", mock.Anything, userID, normalizedParams).Return(expectedResult, tt.mockErr)

			result, err := service.List(context.Background(), userID, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, PaginatedResult[domain.Portrait]{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPortraitService_GetLatest(t *testing.T) {
	userID := uuid.New()

	expectedPortrait := domain.Portrait{
		ID:             uuid.New(),
		UserID:         userID,
		StoragePath:    "latest-portrait.jpg",
		PortraitYear:   2024,
		IsManualUpload: true,
		CapturedAt:     time.Now(),
	}

	tests := []struct {
		name     string
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "success",
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "not found",
			mockErr: domain.ErrNotFound,
			wantErr: true,
		},
		{
			name:    "repo error",
			mockErr: errors.New("repo error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockPortraitRepository{}
			service := NewPortraitService(mockRepo)

			mockRepo.On("GetLatestByUserID", mock.Anything, userID).Return(expectedPortrait, tt.mockErr)

			portrait, err := service.GetLatest(context.Background(), userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, domain.Portrait{}, portrait)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedPortrait, portrait)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPortraitService_Delete(t *testing.T) {
	userID := uuid.New()
	portraitID := uuid.New()

	tests := []struct {
		name     string
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "success",
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "portrait not found",
			mockErr: domain.ErrNotFound,
			wantErr: true,
		},
		{
			name:    "repo error",
			mockErr: errors.New("repo error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockPortraitRepository{}
			service := NewPortraitService(mockRepo)

			mockRepo.On("Delete", mock.Anything, portraitID).Return(tt.mockErr)

			err := service.Delete(context.Background(), userID, portraitID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}