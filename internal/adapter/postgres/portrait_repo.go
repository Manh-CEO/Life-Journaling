package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/life-journaling/core/internal/domain"
	"github.com/life-journaling/core/internal/usecase"
)

// PortraitRepository implements IPortraitRepository using PostgreSQL.
type PortraitRepository struct {
	pool *pgxpool.Pool
}

// NewPortraitRepository creates a new PortraitRepository.
func NewPortraitRepository(pool *pgxpool.Pool) *PortraitRepository {
	return &PortraitRepository{pool: pool}
}

// Create inserts a new portrait.
func (r *PortraitRepository) Create(ctx context.Context, portrait domain.Portrait) (domain.Portrait, error) {
	var created domain.Portrait
	err := r.pool.QueryRow(ctx, `
		INSERT INTO portraits (user_id, storage_path, portrait_year, is_manual_upload, captured_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, storage_path, portrait_year, is_manual_upload, captured_at
	`, portrait.UserID, portrait.StoragePath, portrait.PortraitYear,
		portrait.IsManualUpload, portrait.CapturedAt,
	).Scan(
		&created.ID, &created.UserID, &created.StoragePath,
		&created.PortraitYear, &created.IsManualUpload, &created.CapturedAt,
	)
	if err != nil {
		return domain.Portrait{}, fmt.Errorf("inserting portrait: %w", err)
	}
	return created, nil
}

// GetByUserID retrieves portraits for a user with pagination.
func (r *PortraitRepository) GetByUserID(ctx context.Context, userID uuid.UUID, params usecase.PaginationParams) (usecase.PaginatedResult[domain.Portrait], error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM portraits WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return usecase.PaginatedResult[domain.Portrait]{}, fmt.Errorf("counting portraits: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, storage_path, portrait_year, is_manual_upload, captured_at
		FROM portraits
		WHERE user_id = $1
		ORDER BY portrait_year DESC, captured_at DESC
		LIMIT $2 OFFSET $3
	`, userID, params.Limit, params.Offset)
	if err != nil {
		return usecase.PaginatedResult[domain.Portrait]{}, fmt.Errorf("querying portraits: %w", err)
	}
	defer rows.Close()

	var portraits []domain.Portrait
	for rows.Next() {
		var p domain.Portrait
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.StoragePath,
			&p.PortraitYear, &p.IsManualUpload, &p.CapturedAt,
		); err != nil {
			return usecase.PaginatedResult[domain.Portrait]{}, fmt.Errorf("scanning portrait: %w", err)
		}
		portraits = append(portraits, p)
	}
	if err := rows.Err(); err != nil {
		return usecase.PaginatedResult[domain.Portrait]{}, fmt.Errorf("iterating portraits: %w", err)
	}

	return usecase.PaginatedResult[domain.Portrait]{
		Items: portraits,
		Total: total,
	}, nil
}

// GetLatestByUserID retrieves the latest portrait for a user.
func (r *PortraitRepository) GetLatestByUserID(ctx context.Context, userID uuid.UUID) (domain.Portrait, error) {
	var p domain.Portrait
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, storage_path, portrait_year, is_manual_upload, captured_at
		FROM portraits
		WHERE user_id = $1
		ORDER BY portrait_year DESC, captured_at DESC
		LIMIT 1
	`, userID).Scan(
		&p.ID, &p.UserID, &p.StoragePath,
		&p.PortraitYear, &p.IsManualUpload, &p.CapturedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Portrait{}, domain.ErrNotFound
		}
		return domain.Portrait{}, fmt.Errorf("querying latest portrait: %w", err)
	}
	return p, nil
}

// Delete removes a portrait by ID.
func (r *PortraitRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM portraits WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting portrait: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
