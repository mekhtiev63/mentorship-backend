package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/user/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// UserRepo implements domain.UserRepository.
type UserRepo struct {
	pool *postgres.Pool
}

// NewUserRepo creates UserRepo.
func NewUserRepo(pool *postgres.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// Create inserts a new user with password hash.
func (r *UserRepo) Create(ctx context.Context, email domain.Email, passwordHash string, status domain.UserStatus) (domain.User, error) {
	const q = `
		INSERT INTO users (email, password_hash, status)
		VALUES ($1, $2, $3)
		RETURNING id, email, status, created_at, updated_at, deleted_at
	`
	row := r.pool.QueryRow(ctx, q, string(email), passwordHash, string(status))
	user, err := scanUser(row)
	if isUniqueViolation(err) {
		return domain.User{}, domain.ErrEmailTaken
	}
	return user, err
}

// UpdateStatus updates user status.
func (r *UserRepo) UpdateStatus(ctx context.Context, id domain.UserID, status domain.UserStatus) error {
	const q = `UPDATE users SET status = $2, updated_at = now() WHERE id = $1 AND deleted_at IS NULL`
	ct, err := r.pool.Exec(ctx, q, string(id), string(status))
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SoftDelete sets deleted_at.
func (r *UserRepo) SoftDelete(ctx context.Context, id domain.UserID) error {
	const q = `UPDATE users SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL`
	ct, err := r.pool.Exec(ctx, q, string(id))
	if err != nil {
		return fmt.Errorf("soft delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// FindByID loads user by id.
func (r *UserRepo) FindByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	const q = `
		SELECT id, email, status, created_at, updated_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`
	return scanUser(r.pool.QueryRow(ctx, q, string(id)))
}

// FindByEmail loads user by email.
func (r *UserRepo) FindByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	const q = `
		SELECT id, email, status, created_at, updated_at, deleted_at
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`
	return scanUser(r.pool.QueryRow(ctx, q, string(email)))
}

// ExistsActive reports whether user exists and is not deleted.
func (r *UserRepo) ExistsActive(ctx context.Context, id domain.UserID) (bool, error) {
	const q = `SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL`
	var one int
	err := r.pool.QueryRow(ctx, q, string(id)).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// List returns paginated users.
func (r *UserRepo) List(ctx context.Context, filter domain.UserListFilter, page pagination.Params) ([]domain.User, int64, error) {
	where := `WHERE deleted_at IS NULL`
	args := []any{}
	argn := 1

	if filter.EmailPrefix != "" {
		where += fmt.Sprintf(` AND email::text ILIKE $%d`, argn)
		args = append(args, filter.EmailPrefix+"%")
		argn++
	}
	if filter.Status != nil {
		where += fmt.Sprintf(` AND status = $%d`, argn)
		args = append(args, string(*filter.Status))
		argn++
	}
	if filter.Role != nil {
		where += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = users.id AND ur.role = $%d)`, argn)
		args = append(args, string(*filter.Role))
		argn++
	}

	countQ := `SELECT COUNT(*) FROM users ` + where
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	listQ := fmt.Sprintf(`
		SELECT id, email, status, created_at, updated_at, deleted_at
		FROM users %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argn, argn+1)
	args = append(args, page.PageSize, page.Offset)

	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func scanUser(row pgx.Row) (domain.User, error) {
	var (
		id, email, status string
		createdAt, updatedAt time.Time
		deletedAt *time.Time
	)
	err := row.Scan(&id, &email, &status, &createdAt, &updatedAt, &deletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("scan user: %w", err)
	}
	return domain.User{
		ID:        domain.UserID(id),
		Email:     domain.Email(email),
		Status:    domain.UserStatus(status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
