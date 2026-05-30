package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/identity/domain"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
)

// CredentialsRepo loads users for authentication.
type CredentialsRepo struct {
	pool *postgres.Pool
}

// NewCredentialsRepo creates a CredentialsRepo.
func NewCredentialsRepo(pool *postgres.Pool) *CredentialsRepo {
	return &CredentialsRepo{pool: pool}
}

// FindByEmail returns credentials by email.
func (r *CredentialsRepo) FindByEmail(ctx context.Context, email domain.Email) (domain.Credentials, error) {
	const q = `
		SELECT id, email, password_hash, status, deleted_at IS NOT NULL
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`
	return r.scanOne(ctx, q, email.String(), domain.ErrInvalidCredentials)
}

// FindByID returns credentials by user id.
func (r *CredentialsRepo) FindByID(ctx context.Context, userID domain.UserID) (domain.Credentials, error) {
	const q = `
		SELECT id, email, password_hash, status, deleted_at IS NOT NULL
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.scanOne(ctx, q, userID.String(), domain.ErrAccountInactive)
}

func (r *CredentialsRepo) scanOne(ctx context.Context, query string, arg any, notFound error) (domain.Credentials, error) {
	var (
		id           string
		email        string
		passwordHash string
		status       string
		deleted      bool
	)

	err := r.pool.QueryRow(ctx, query, arg).Scan(&id, &email, &passwordHash, &status, &deleted)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Credentials{}, notFound
	}
	if err != nil {
		return domain.Credentials{}, fmt.Errorf("query credentials: %w", err)
	}

	uid, err := domain.ParseUserID(id)
	if err != nil {
		return domain.Credentials{}, err
	}

	return domain.Credentials{
		UserID:       uid,
		Email:        domain.Email(email),
		PasswordHash: domain.PasswordHash(passwordHash),
		Account: domain.AccountState{
			Status:  domain.AccountStatus(status),
			Deleted: deleted,
		},
	}, nil
}
