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

// MockMemoryRepository is a mock implementation of IMemoryRepository
type MockMemoryRepository struct {
	mock.Mock
}

func (m *MockMemoryRepository) Create(ctx context.Context, memory domain.Memory) (domain.Memory, error) {
	args := m.Called(ctx, memory)
	return args.Get(0).(domain.Memory), args.Error(1)
}

func (m *MockMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Memory, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Memory), args.Error(1)
}

func (m *MockMemoryRepository) GetByUserID(ctx context.Context, userID uuid.UUID, params PaginationParams) (PaginatedResult[domain.Memory], error) {
	args := m.Called(ctx, userID, params)
	return args.Get(0).(PaginatedResult[domain.Memory]), args.Error(1)
}

func (m *MockMemoryRepository) Update(ctx context.Context, memory domain.Memory) (domain.Memory, error) {
	args := m.Called(ctx, memory)
	return args.Get(0).(domain.Memory), args.Error(1)
}

func (m *MockMemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestMemoryService_Create(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name     string
		input    CreateMemoryInput
		mockErr  error
		wantErr  bool
		errType  error
	}{
		{
			name: "success with all fields",
			input: CreateMemoryInput{
				EntryDate: "2024-01-15",
				Location:  "Home",
				Content:   "Had a great day",
				Sentiment: "positive",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "success with minimal fields",
			input: CreateMemoryInput{
				EntryDate: "2024-01-15",
				Content:   "Simple note",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "invalid date format",
			input: CreateMemoryInput{
				EntryDate: "invalid-date",
				Content:   "Test",
			},
			wantErr: true,
			errType: domain.ErrValidation,
		},
		{
			name: "empty content",
			input: CreateMemoryInput{
				EntryDate: "2024-01-15",
				Content:   "",
			},
			wantErr: true,
			errType: domain.ErrValidation,
		},
		{
			name: "repo create fails",
			input: CreateMemoryInput{
				EntryDate: "2024-01-15",
				Content:   "Test",
			},
			mockErr: errors.New("repo error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockMemoryRepository{}
			service := NewMemoryService(mockRepo)

			if !tt.wantErr || tt.errType != domain.ErrValidation {
				// Mock the repo call for non-validation errors
				expectedMemory := domain.Memory{
					UserID:        userID,
					EntryDate:     parseDate(t, tt.input.EntryDate),
					Location:      tt.input.Location,
					Content:       tt.input.Content,
					Sentiment:     tt.input.Sentiment,
					IsManualEntry: true,
				}
				if expectedMemory.Sentiment == "" {
					expectedMemory.Sentiment = "neutral"
				}

				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(m domain.Memory) bool {
					return m.UserID == userID && m.Content == tt.input.Content && m.IsManualEntry
				})).Return(expectedMemory, tt.mockErr)
			}

			memory, err := service.Create(context.Background(), userID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.True(t, errors.Is(err, tt.errType))
				}
				assert.Equal(t, domain.Memory{}, memory)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, userID, memory.UserID)
				assert.Equal(t, tt.input.Content, memory.Content)
				assert.True(t, memory.IsManualEntry)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMemoryService_GetByID(t *testing.T) {
	userID := uuid.New()
	memoryID := uuid.New()
	otherUserID := uuid.New()

	ownedMemory := domain.Memory{
		ID:     memoryID,
		UserID: userID,
		Content: "My memory",
	}

	unownedMemory := domain.Memory{
		ID:     memoryID,
		UserID: otherUserID,
		Content: "Someone else's memory",
	}

	tests := []struct {
		name      string
		requestID uuid.UUID
		mockMemory domain.Memory
		mockErr    error
		wantErr    bool
		errType    error
	}{
		{
			name:       "success - owned memory",
			requestID:  userID,
			mockMemory: ownedMemory,
			mockErr:    nil,
			wantErr:    false,
		},
		{
			name:       "forbidden - unowned memory",
			requestID:  userID,
			mockMemory: unownedMemory,
			mockErr:    nil,
			wantErr:    true,
			errType:    domain.ErrForbidden,
		},
		{
			name:       "memory not found",
			requestID:  userID,
			mockMemory: domain.Memory{},
			mockErr:    domain.ErrNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockMemoryRepository{}
			service := NewMemoryService(mockRepo)

			mockRepo.On("GetByID", mock.Anything, memoryID).Return(tt.mockMemory, tt.mockErr)

			memory, err := service.GetByID(context.Background(), tt.requestID, memoryID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.Equal(t, tt.errType, err)
				}
				assert.Equal(t, domain.Memory{}, memory)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockMemory, memory)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMemoryService_List(t *testing.T) {
	userID := uuid.New()
	params := PaginationParams{Limit: 10, Offset: 0}

	expectedResult := PaginatedResult[domain.Memory]{
		Items: []domain.Memory{
			{ID: uuid.New(), UserID: userID, Content: "Memory 1"},
			{ID: uuid.New(), UserID: userID, Content: "Memory 2"},
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
			mockRepo := &MockMemoryRepository{}
			service := NewMemoryService(mockRepo)

			normalizedParams := normalizePagination(tt.params)
			mockRepo.On("GetByUserID", mock.Anything, userID, normalizedParams).Return(expectedResult, tt.mockErr)

			result, err := service.List(context.Background(), userID, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, PaginatedResult[domain.Memory]{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMemoryService_Update(t *testing.T) {
	userID := uuid.New()
	memoryID := uuid.New()
	otherUserID := uuid.New()

	ownedMemory := domain.Memory{
		ID:            memoryID,
		UserID:        userID,
		EntryDate:     parseDate(t, "2024-01-15"),
		Location:      "Home",
		Content:       "Original content",
		Sentiment:     "neutral",
		IsManualEntry: true,
	}

	unownedMemory := domain.Memory{
		ID:     memoryID,
		UserID: otherUserID,
	}

	tests := []struct {
		name      string
		requestID uuid.UUID
		input     UpdateMemoryInput
		mockMemory domain.Memory
		mockErr    error
		wantErr    bool
		errType    error
	}{
		{
			name:       "success - update all fields",
			requestID:  userID,
			mockMemory: ownedMemory,
			input: UpdateMemoryInput{
				EntryDate: &[]string{"2024-01-16"}[0],
				Location:  &[]string{"Office"}[0],
				Content:   &[]string{"Updated content"}[0],
				Sentiment: &[]string{"positive"}[0],
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:       "success - partial update",
			requestID:  userID,
			mockMemory: ownedMemory,
			input: UpdateMemoryInput{
				Content: &[]string{"New content only"}[0],
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:       "forbidden - unowned memory",
			requestID:  userID,
			mockMemory: unownedMemory,
			input: UpdateMemoryInput{
				Content: &[]string{"Updated"}[0],
			},
			wantErr: true,
			errType: domain.ErrForbidden,
		},
		{
			name:       "invalid date format",
			requestID:  userID,
			mockMemory: ownedMemory,
			input: UpdateMemoryInput{
				EntryDate: &[]string{"invalid-date"}[0],
			},
			wantErr: true,
			errType: domain.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockMemoryRepository{}
			service := NewMemoryService(mockRepo)

			if !tt.wantErr || tt.errType != domain.ErrValidation {
				mockRepo.On("GetByID", mock.Anything, memoryID).Return(tt.mockMemory, nil)

				if !tt.wantErr {
					mockRepo.On("Update", mock.Anything, mock.Anything).Return(tt.mockMemory, tt.mockErr)
				}
			} else {
				mockRepo.On("GetByID", mock.Anything, memoryID).Return(tt.mockMemory, nil)
			}

			memory, err := service.Update(context.Background(), userID, memoryID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.True(t, errors.Is(err, tt.errType))
				}
				assert.Equal(t, domain.Memory{}, memory)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, domain.Memory{}, memory)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMemoryService_Delete(t *testing.T) {
	userID := uuid.New()
	memoryID := uuid.New()
	otherUserID := uuid.New()

	ownedMemory := domain.Memory{
		ID:     memoryID,
		UserID: userID,
	}

	unownedMemory := domain.Memory{
		ID:     memoryID,
		UserID: otherUserID,
	}

	tests := []struct {
		name      string
		requestID uuid.UUID
		mockMemory domain.Memory
		mockErr    error
		wantErr    bool
		errType    error
	}{
		{
			name:       "success",
			requestID:  userID,
			mockMemory: ownedMemory,
			mockErr:    nil,
			wantErr:    false,
		},
		{
			name:       "forbidden - unowned memory",
			requestID:  userID,
			mockMemory: unownedMemory,
			wantErr:    true,
			errType:    domain.ErrForbidden,
		},
		{
			name:       "memory not found",
			requestID:  userID,
			mockMemory: domain.Memory{},
			mockErr:    domain.ErrNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockMemoryRepository{}
			service := NewMemoryService(mockRepo)

			if !tt.wantErr {
				mockRepo.On("GetByID", mock.Anything, memoryID).Return(tt.mockMemory, nil)
				mockRepo.On("Delete", mock.Anything, memoryID).Return(tt.mockErr)
			} else if tt.errType == domain.ErrForbidden {
				mockRepo.On("GetByID", mock.Anything, memoryID).Return(tt.mockMemory, nil)
			} else {
				mockRepo.On("GetByID", mock.Anything, memoryID).Return(tt.mockMemory, tt.mockErr)
			}

			err := service.Delete(context.Background(), userID, memoryID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.Equal(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNormalizePagination(t *testing.T) {
	tests := []struct {
		name     string
		input    PaginationParams
		expected PaginationParams
	}{
		{
			name:     "default values",
			input:    PaginationParams{},
			expected: PaginationParams{Limit: 20, Offset: 0},
		},
		{
			name:     "negative offset",
			input:    PaginationParams{Limit: 10, Offset: -5},
			expected: PaginationParams{Limit: 10, Offset: 0},
		},
		{
			name:     "limit too high",
			input:    PaginationParams{Limit: 200, Offset: 0},
			expected: PaginationParams{Limit: 100, Offset: 0},
		},
		{
			name:     "negative limit",
			input:    PaginationParams{Limit: -5, Offset: 0},
			expected: PaginationParams{Limit: 20, Offset: 0},
		},
		{
			name:     "valid values",
			input:    PaginationParams{Limit: 50, Offset: 100},
			expected: PaginationParams{Limit: 50, Offset: 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePagination(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to parse date in tests
func parseDate(t *testing.T, dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	d, err := time.Parse("2006-01-02", dateStr)
	assert.NoError(t, err)
	return d
}