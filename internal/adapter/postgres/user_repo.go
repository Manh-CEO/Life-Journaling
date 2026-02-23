package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/life-journaling/core/internal/domain"
	"github.com/life-journaling/core/internal/usecase"
)

// UserRepository implements IUserRepository using PostgreSQL.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var user domain.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, timezone, anchor_date, prompt_day_of_week, prompt_hour, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Email, &user.Timezone, &user.AnchorDate,
		&user.PromptDayOfWeek, &user.PromptHour, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("querying user by id: %w", err)
	}
	return user, nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, timezone, anchor_date, prompt_day_of_week, prompt_hour, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Email, &user.Timezone, &user.AnchorDate,
		&user.PromptDayOfWeek, &user.PromptHour, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("querying user by email: %w", err)
	}
	return user, nil
}

// Create inserts a new user.
func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	var created domain.User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (id, email, timezone, anchor_date, prompt_day_of_week, prompt_hour)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, timezone, anchor_date, prompt_day_of_week, prompt_hour, created_at, updated_at
	`, user.ID, user.Email, user.Timezone, user.AnchorDate,
		user.PromptDayOfWeek, user.PromptHour,
	).Scan(
		&created.ID, &created.Email, &created.Timezone, &created.AnchorDate,
		&created.PromptDayOfWeek, &created.PromptHour, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("inserting user: %w", err)
	}
	return created, nil
}

// UpdateProfile updates a user's profile fields.
func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, update usecase.UserProfileUpdate) (domain.User, error) {
	// Build dynamic update query
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if update.Timezone != nil {
		setClauses = append(setClauses, fmt.Sprintf("timezone = $%d", argIdx))
		args = append(args, *update.Timezone)
		argIdx++
	}
	if update.AnchorDate != nil {
		parsed, err := time.Parse("2006-01-02", *update.AnchorDate)
		if err != nil {
			return domain.User{}, domain.NewDomainError(domain.ErrValidation, "invalid anchor_date format")
		}
		setClauses = append(setClauses, fmt.Sprintf("anchor_date = $%d", argIdx))
		args = append(args, parsed)
		argIdx++
	}
	if update.PromptDayOfWeek != nil {
		setClauses = append(setClauses, fmt.Sprintf("prompt_day_of_week = $%d", argIdx))
		args = append(args, *update.PromptDayOfWeek)
		argIdx++
	}
	if update.PromptHour != nil {
		setClauses = append(setClauses, fmt.Sprintf("prompt_hour = $%d", argIdx))
		args = append(args, *update.PromptHour)
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	query := fmt.Sprintf(`
		UPDATE users SET %s
		WHERE id = $%d
		RETURNING id, email, timezone, anchor_date, prompt_day_of_week, prompt_hour, created_at, updated_at
	`, joinStrings(setClauses, ", "), argIdx)
	args = append(args, id)

	var user domain.User
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Email, &user.Timezone, &user.AnchorDate,
		&user.PromptDayOfWeek, &user.PromptHour, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("updating user profile: %w", err)
	}
	return user, nil
}

// GetUsersForPrompt retrieves users scheduled for prompts at the given day/hour.
func (r *UserRepository) GetUsersForPrompt(ctx context.Context, dayOfWeek int, hour int) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, email, timezone, anchor_date, prompt_day_of_week, prompt_hour, created_at, updated_at
		FROM users
		WHERE prompt_day_of_week = $1 AND prompt_hour = $2
	`, dayOfWeek, hour)
	if err != nil {
		return nil, fmt.Errorf("querying users for prompt: %w", err)
	}
	defer rows.Close()

	return scanUsers(rows)
}

// GetUsersForAnchorDate retrieves users whose anchor date matches the given month/day.
func (r *UserRepository) GetUsersForAnchorDate(ctx context.Context, month int, day int) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, email, timezone, anchor_date, prompt_day_of_week, prompt_hour, created_at, updated_at
		FROM users
		WHERE anchor_date IS NOT NULL
		  AND EXTRACT(MONTH FROM anchor_date) = $1
		  AND EXTRACT(DAY FROM anchor_date) = $2
	`, month, day)
	if err != nil {
		return nil, fmt.Errorf("querying users for anchor date: %w", err)
	}
	defer rows.Close()

	return scanUsers(rows)
}

func scanUsers(rows pgx.Rows) ([]domain.User, error) {
	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID, &user.Email, &user.Timezone, &user.AnchorDate,
			&user.PromptDayOfWeek, &user.PromptHour, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning user row: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user rows: %w", err)
	}
	return users, nil
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
