package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/life-journaling/core/internal/domain"
)

// IngestionService processes pending engagement logs into memories.
type IngestionService struct {
	engagementRepo IEngagementLogRepository
	memoryRepo     IMemoryRepository
	llmProvider    ILLMProvider
}

// NewIngestionService creates a new IngestionService.
func NewIngestionService(
	engagementRepo IEngagementLogRepository,
	memoryRepo IMemoryRepository,
	llmProvider ILLMProvider,
) *IngestionService {
	return &IngestionService{
		engagementRepo: engagementRepo,
		memoryRepo:     memoryRepo,
		llmProvider:    llmProvider,
	}
}

// IngestEmail stores the raw email as a pending engagement log and creates
// a simple memory from it. V1 saves the raw text directly; full AI parsing
// is deferred to V2.
func (s *IngestionService) IngestEmail(ctx context.Context, fromEmail string, rawText string, userRepo IUserRepository) error {
	user, err := userRepo.GetByEmail(ctx, fromEmail)
	if err != nil {
		return fmt.Errorf("looking up user by email %q: %w", fromEmail, err)
	}

	// Create engagement log
	log := domain.EngagementLog{
		UserID:       user.ID,
		RawEmailText: rawText,
		Status:       domain.EngagementStatusPending,
	}

	created, err := s.engagementRepo.Create(ctx, log)
	if err != nil {
		return fmt.Errorf("creating engagement log: %w", err)
	}

	// V1: Save raw text as memory directly (no LLM parsing)
	memory := domain.Memory{
		UserID:        user.ID,
		EntryDate:     time.Now().UTC(),
		Content:       rawText,
		Sentiment:     "neutral",
		IsManualEntry: false,
	}

	if _, err := s.memoryRepo.Create(ctx, memory); err != nil {
		// Mark engagement log as failed
		if updateErr := s.engagementRepo.UpdateStatus(ctx, created.ID, domain.EngagementStatusFailed); updateErr != nil {
			slog.Error("failed to update engagement log status", "error", updateErr)
		}
		return fmt.Errorf("creating memory from email: %w", err)
	}

	// Mark engagement log as completed
	if err := s.engagementRepo.UpdateStatus(ctx, created.ID, domain.EngagementStatusCompleted); err != nil {
		slog.Error("failed to update engagement log status", "error", err)
	}

	slog.Info("email ingested successfully", "user_id", user.ID, "engagement_log_id", created.ID)
	return nil
}
