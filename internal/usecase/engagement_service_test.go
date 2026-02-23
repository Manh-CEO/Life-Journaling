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

// MockEngagementLogRepository is a mock implementation of IEngagementLogRepository
type MockEngagementLogRepository struct {
	mock.Mock
}

func (m *MockEngagementLogRepository) Create(ctx context.Context, log domain.EngagementLog) (domain.EngagementLog, error) {
	args := m.Called(ctx, log)
	return args.Get(0).(domain.EngagementLog), args.Error(1)
}

func (m *MockEngagementLogRepository) GetPendingByUserID(ctx context.Context, userID uuid.UUID) ([]domain.EngagementLog, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.EngagementLog), args.Error(1)
}

func (m *MockEngagementLogRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// MockEmailProvider is a mock implementation of IEmailProvider
type MockEmailProvider struct {
	mock.Mock
}

func (m *MockEmailProvider) SendPrompt(ctx context.Context, toEmail string, subject string, body string) error {
	args := m.Called(ctx, toEmail, subject, body)
	return args.Error(0)
}

func TestEngagementService_SendHourlyPrompts(t *testing.T) {
	// Mock time: Monday (1), 08:00 UTC

	tests := []struct {
		name              string
		mockUsers         []domain.User
		mockUserErr       error
		mockEmailErr      error
		expectedEmailCalls int
		wantErr           bool
	}{
		{
			name: "success - multiple users",
			mockUsers: []domain.User{
				{ID: uuid.New(), Email: "user1@example.com", PromptDayOfWeek: 1, PromptHour: 8},
				{ID: uuid.New(), Email: "user2@example.com", PromptDayOfWeek: 1, PromptHour: 8},
			},
			mockUserErr:       nil,
			mockEmailErr:      nil,
			expectedEmailCalls: 2,
			wantErr:           false,
		},
		{
			name:             "success - no users",
			mockUsers:        []domain.User{},
			mockUserErr:      nil,
			mockEmailErr:     nil,
			expectedEmailCalls: 0,
			wantErr:          false,
		},
		{
			name: "user repo error",
			mockUsers:        nil,
			mockUserErr:      errors.New("repo error"),
			expectedEmailCalls: 0,
			wantErr:          true,
		},
		{
			name: "email send failure - continues processing",
			mockUsers: []domain.User{
				{ID: uuid.New(), Email: "user1@example.com", PromptDayOfWeek: 1, PromptHour: 8},
				{ID: uuid.New(), Email: "user2@example.com", PromptDayOfWeek: 1, PromptHour: 8},
			},
			mockUserErr:       nil,
			mockEmailErr:      errors.New("email send failed"),
			expectedEmailCalls: 2, // Both calls attempted
			wantErr:           false, // Service doesn't fail on email errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &MockUserRepository{}
			mockEngagementRepo := &MockEngagementLogRepository{}
			mockEmailProvider := &MockEmailProvider{}

			service := NewEngagementService(mockUserRepo, mockEngagementRepo, mockEmailProvider)

			mockUserRepo.On("GetUsersForPrompt", mock.Anything, mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(tt.mockUsers, tt.mockUserErr)

			// Set up email expectations
			for _, user := range tt.mockUsers {
				mockEmailProvider.On("SendPrompt", mock.Anything, user.Email, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(tt.mockEmailErr)
			}

			err := service.SendHourlyPrompts(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockEmailProvider.AssertNumberOfCalls(t, "SendPrompt", tt.expectedEmailCalls)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestEngagementService_SendAnchorDateEmails(t *testing.T) {
	// Mock time: January 15

	tests := []struct {
		name              string
		mockUsers         []domain.User
		mockUserErr       error
		mockEmailErr      error
		expectedEmailCalls int
		wantErr           bool
	}{
		{
			name: "success - multiple users",
			mockUsers: []domain.User{
				{ID: uuid.New(), Email: "user1@example.com"},
				{ID: uuid.New(), Email: "user2@example.com"},
			},
			mockUserErr:       nil,
			mockEmailErr:      nil,
			expectedEmailCalls: 2,
			wantErr:           false,
		},
		{
			name:             "success - no users",
			mockUsers:        []domain.User{},
			mockUserErr:      nil,
			mockEmailErr:     nil,
			expectedEmailCalls: 0,
			wantErr:          false,
		},
		{
			name: "user repo error",
			mockUsers:        nil,
			mockUserErr:      errors.New("repo error"),
			expectedEmailCalls: 0,
			wantErr:          true,
		},
		{
			name: "email send failure - continues processing",
			mockUsers: []domain.User{
				{ID: uuid.New(), Email: "user1@example.com"},
				{ID: uuid.New(), Email: "user2@example.com"},
			},
			mockUserErr:       nil,
			mockEmailErr:      errors.New("email send failed"),
			expectedEmailCalls: 2,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &MockUserRepository{}
			mockEngagementRepo := &MockEngagementLogRepository{}
			mockEmailProvider := &MockEmailProvider{}

			service := NewEngagementService(mockUserRepo, mockEngagementRepo, mockEmailProvider)

			mockUserRepo.On("GetUsersForAnchorDate", mock.Anything, mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(tt.mockUsers, tt.mockUserErr)

			// Set up email expectations
			for _, user := range tt.mockUsers {
				mockEmailProvider.On("SendPrompt", mock.Anything, user.Email, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(tt.mockEmailErr)
			}

			err := service.SendAnchorDateEmails(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockEmailProvider.AssertNumberOfCalls(t, "SendPrompt", tt.expectedEmailCalls)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestEngagementService_LogInboundEmail(t *testing.T) {
	userID := uuid.New()
	fromEmail := "test@example.com"
	rawText := "This is my weekly reflection..."

	user := domain.User{
		ID:    userID,
		Email: fromEmail,
	}

	expectedLog := domain.EngagementLog{
		UserID:       userID,
		RawEmailText: rawText,
		Status:       domain.EngagementStatusPending,
	}

	tests := []struct {
		name         string
		fromEmail    string
		rawText      string
		mockUser     domain.User
		mockUserErr  error
		mockCreateErr error
		wantErr      bool
	}{
		{
			name:          "success",
			fromEmail:     fromEmail,
			rawText:       rawText,
			mockUser:      user,
			mockUserErr:   nil,
			mockCreateErr: nil,
			wantErr:       false,
		},
		{
			name:         "user not found",
			fromEmail:    fromEmail,
			rawText:      rawText,
			mockUser:     domain.User{},
			mockUserErr:  domain.ErrNotFound,
			wantErr:      true,
		},
		{
			name:          "user lookup error",
			fromEmail:     fromEmail,
			rawText:       rawText,
			mockUser:      domain.User{},
			mockUserErr:   errors.New("db error"),
			wantErr:       true,
		},
		{
			name:          "engagement log creation error",
			fromEmail:     fromEmail,
			rawText:       rawText,
			mockUser:      user,
			mockUserErr:   nil,
			mockCreateErr: errors.New("create error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &MockUserRepository{}
			mockEngagementRepo := &MockEngagementLogRepository{}
			mockEmailProvider := &MockEmailProvider{}

			service := NewEngagementService(mockUserRepo, mockEngagementRepo, mockEmailProvider)

			mockUserRepo.On("GetByEmail", mock.Anything, tt.fromEmail).Return(tt.mockUser, tt.mockUserErr)

			if tt.mockUserErr == nil {
				mockEngagementRepo.On("Create", mock.Anything, mock.MatchedBy(func(log domain.EngagementLog) bool {
					return log.UserID == tt.mockUser.ID && log.RawEmailText == tt.rawText && log.Status == domain.EngagementStatusPending
				})).Return(expectedLog, tt.mockCreateErr)
			}

			log, err := service.LogInboundEmail(context.Background(), tt.fromEmail, tt.rawText)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, domain.EngagementLog{}, log)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedLog.UserID, log.UserID)
				assert.Equal(t, expectedLog.RawEmailText, log.RawEmailText)
				assert.Equal(t, expectedLog.Status, log.Status)
			}

			mockUserRepo.AssertExpectations(t)
			mockEngagementRepo.AssertExpectations(t)
		})
	}
}