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

// MemoryRepository implements IMemoryRepository using PostgreSQL.
type MemoryRepository struct {
	pool *pgxpool.Pool
}

// NewMemoryRepository creates a new MemoryRepository.
func NewMemoryRepository(pool *pgxpool.Pool) *MemoryRepository {
	return &MemoryRepository{pool: pool}
}

// Create inserts a new memory.
func (r *MemoryRepository) Create(ctx context.Context, memory domain.Memory) (domain.Memory, error) {
	var created domain.Memory
	err := r.pool.QueryRow(ctx, `
		INSERT INTO memories (user_id, entry_date, location, content, sentiment, is_manual_entry)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, entry_date, location, content, sentiment, is_manual_entry, created_at, updated_at
	`, memory.UserID, memory.EntryDate, memory.Location, memory.Content,
		memory.Sentiment, memory.IsManualEntry,
	).Scan(
		&created.ID, &created.UserID, &created.EntryDate, &created.Location,
		&created.Content, &created.Sentiment, &created.IsManualEntry,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return domain.Memory{}, fmt.Errorf("inserting memory: %w", err)
	}
	return created, nil
}

// GetByID retrieves a memory by ID.
func (r *MemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Memory, error) {
	var m domain.Memory
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, entry_date, location, content, sentiment, is_manual_entry, created_at, updated_at
		FROM memories WHERE id = $1
	`, id).Scan(
		&m.ID, &m.UserID, &m.EntryDate, &m.Location,
		&m.Content, &m.Sentiment, &m.IsManualEntry,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Memory{}, domain.ErrNotFound
		}
		return domain.Memory{}, fmt.Errorf("querying memory by id: %w", err)
	}
	return m, nil
}

// GetByUserID retrieves memories for a user with pagination.
func (r *MemoryRepository) GetByUserID(ctx context.Context, userID uuid.UUID, params usecase.PaginationParams) (usecase.PaginatedResult[domain.Memory], error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM memories WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return usecase.PaginatedResult[domain.Memory]{}, fmt.Errorf("counting memories: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, entry_date, location, content, sentiment, is_manual_entry, created_at, updated_at
		FROM memories
		WHERE user_id = $1
		ORDER BY entry_date DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, params.Limit, params.Offset)
	if err != nil {
		return usecase.PaginatedResult[domain.Memory]{}, fmt.Errorf("querying memories: %w", err)
	}
	defer rows.Close()

	var memories []domain.Memory
	for rows.Next() {
		var m domain.Memory
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.EntryDate, &m.Location,
			&m.Content, &m.Sentiment, &m.IsManualEntry,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return usecase.PaginatedResult[domain.Memory]{}, fmt.Errorf("scanning memory: %w", err)
		}
		memories = append(memories, m)
	}
	if err := rows.Err(); err != nil {
		return usecase.PaginatedResult[domain.Memory]{}, fmt.Errorf("iterating memories: %w", err)
	}

	return usecase.PaginatedResult[domain.Memory]{
		Items: memories,
		Total: total,
	}, nil
}

// Update updates a memory.
func (r *MemoryRepository) Update(ctx context.Context, memory domain.Memory) (domain.Memory, error) {
	var updated domain.Memory
	err := r.pool.QueryRow(ctx, `
		UPDATE memories
		SET entry_date = $1, location = $2, content = $3, sentiment = $4
		WHERE id = $5
		RETURNING id, user_id, entry_date, location, content, sentiment, is_manual_entry, created_at, updated_at
	`, memory.EntryDate, memory.Location, memory.Content, memory.Sentiment, memory.ID,
	).Scan(
		&updated.ID, &updated.UserID, &updated.EntryDate, &updated.Location,
		&updated.Content, &updated.Sentiment, &updated.IsManualEntry,
		&updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Memory{}, domain.ErrNotFound
		}
		return domain.Memory{}, fmt.Errorf("updating memory: %w", err)
	}
	return updated, nil
}

// Delete removes a memory by ID.
func (r *MemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM memories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting memory: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
