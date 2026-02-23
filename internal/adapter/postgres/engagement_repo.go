package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/life-journaling/core/internal/domain"
)

// EngagementLogRepository implements IEngagementLogRepository using PostgreSQL.
type EngagementLogRepository struct {
	pool *pgxpool.Pool
}

// NewEngagementLogRepository creates a new EngagementLogRepository.
func NewEngagementLogRepository(pool *pgxpool.Pool) *EngagementLogRepository {
	return &EngagementLogRepository{pool: pool}
}

// Create inserts a new engagement log.
func (r *EngagementLogRepository) Create(ctx context.Context, log domain.EngagementLog) (domain.EngagementLog, error) {
	var created domain.EngagementLog
	err := r.pool.QueryRow(ctx, `
		INSERT INTO engagement_logs (user_id, raw_email_text, status)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, raw_email_text, status, received_at
	`, log.UserID, log.RawEmailText, log.Status,
	).Scan(
		&created.ID, &created.UserID, &created.RawEmailText,
		&created.Status, &created.ReceivedAt,
	)
	if err != nil {
		return domain.EngagementLog{}, fmt.Errorf("inserting engagement log: %w", err)
	}
	return created, nil
}

// GetPendingByUserID retrieves pending engagement logs for a user.
func (r *EngagementLogRepository) GetPendingByUserID(ctx context.Context, userID uuid.UUID) ([]domain.EngagementLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, raw_email_text, status, received_at
		FROM engagement_logs
		WHERE user_id = $1 AND status = 'pending'
		ORDER BY received_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("querying pending engagement logs: %w", err)
	}
	defer rows.Close()

	var logs []domain.EngagementLog
	for rows.Next() {
		var log domain.EngagementLog
		if err := rows.Scan(
			&log.ID, &log.UserID, &log.RawEmailText,
			&log.Status, &log.ReceivedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning engagement log: %w", err)
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating engagement logs: %w", err)
	}
	return logs, nil
}

// UpdateStatus updates the status of an engagement log.
func (r *EngagementLogRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE engagement_logs SET status = $1 WHERE id = $2
	`, status, id)
	if err != nil {
		return fmt.Errorf("updating engagement log status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
