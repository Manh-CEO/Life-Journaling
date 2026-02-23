package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/life-journaling/core/internal/domain"
)

// EngagementService handles sending weekly prompts and anchor-date emails.
type EngagementService struct {
	userRepo       IUserRepository
	engagementRepo IEngagementLogRepository
	emailProvider  IEmailProvider
}

// NewEngagementService creates a new EngagementService.
func NewEngagementService(
	userRepo IUserRepository,
	engagementRepo IEngagementLogRepository,
	emailProvider IEmailProvider,
) *EngagementService {
	return &EngagementService{
		userRepo:       userRepo,
		engagementRepo: engagementRepo,
		emailProvider:  emailProvider,
	}
}

// SendHourlyPrompts finds users whose prompt schedule matches the current
// day-of-week and hour (UTC), then sends each a prompt email.
func (s *EngagementService) SendHourlyPrompts(ctx context.Context) error {
	now := time.Now().UTC()
	dayOfWeek := int(now.Weekday())
	hour := now.Hour()

	users, err := s.userRepo.GetUsersForPrompt(ctx, dayOfWeek, hour)
	if err != nil {
		return fmt.Errorf("fetching users for prompt: %w", err)
	}

	slog.Info("sending hourly prompts", "user_count", len(users), "day", dayOfWeek, "hour", hour)

	for _, user := range users {
		subject := "What's on your mind this week?"
		body := fmt.Sprintf(
			"Hi there!\n\nTake a moment to reflect on your week. "+
				"Simply reply to this email with your thoughts, memories, or anything you'd like to remember.\n\n"+
				"— Life Journaling",
		)

		if err := s.emailProvider.SendPrompt(ctx, user.Email, subject, body); err != nil {
			slog.Error("failed to send prompt email",
				"user_id", user.ID,
				"email", user.Email,
				"error", err,
			)
			continue
		}

		slog.Info("prompt email sent", "user_id", user.ID)
	}

	return nil
}

// SendAnchorDateEmails finds users whose anchor date matches today's month/day
// and sends them a special anniversary email.
func (s *EngagementService) SendAnchorDateEmails(ctx context.Context) error {
	now := time.Now().UTC()
	month := int(now.Month())
	day := now.Day()

	users, err := s.userRepo.GetUsersForAnchorDate(ctx, month, day)
	if err != nil {
		return fmt.Errorf("fetching users for anchor date: %w", err)
	}

	slog.Info("sending anchor date emails", "user_count", len(users), "month", month, "day", day)

	for _, user := range users {
		subject := "It's your anchor date! 🎉"
		body := fmt.Sprintf(
			"Hi there!\n\nToday is your special anchor date. "+
				"Time to capture a portrait and reflect on the past year.\n\n"+
				"Reply to this email with your thoughts!\n\n"+
				"— Life Journaling",
		)

		if err := s.emailProvider.SendPrompt(ctx, user.Email, subject, body); err != nil {
			slog.Error("failed to send anchor date email",
				"user_id", user.ID,
				"email", user.Email,
				"error", err,
			)
			continue
		}

		slog.Info("anchor date email sent", "user_id", user.ID)
	}

	return nil
}

// LogInboundEmail stores a raw inbound email as a pending engagement log.
func (s *EngagementService) LogInboundEmail(ctx context.Context, fromEmail string, rawText string) (domain.EngagementLog, error) {
	user, err := s.userRepo.GetByEmail(ctx, fromEmail)
	if err != nil {
		return domain.EngagementLog{}, fmt.Errorf("looking up user by email %q: %w", fromEmail, err)
	}

	log := domain.EngagementLog{
		UserID:       user.ID,
		RawEmailText: rawText,
		Status:       domain.EngagementStatusPending,
	}

	created, err := s.engagementRepo.Create(ctx, log)
	if err != nil {
		return domain.EngagementLog{}, fmt.Errorf("creating engagement log: %w", err)
	}

	return created, nil
}
