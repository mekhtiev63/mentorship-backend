package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/identity/domain"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
)

// RefreshTokenRepo persists refresh_tokens.
type RefreshTokenRepo struct {
	pool *postgres.Pool
}

// NewRefreshTokenRepo creates a RefreshTokenRepo.
func NewRefreshTokenRepo(pool *postgres.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{pool: pool}
}

// Store inserts a refresh token row.
func (r *RefreshTokenRepo) Store(ctx context.Context, token domain.RefreshToken) error {
	const q = `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.pool.Exec(ctx, q, string(token.ID), token.UserID.String(), token.TokenHash, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

// FindByHash loads a token by hash.
func (r *RefreshTokenRepo) FindByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var (
		id        string
		userID    string
		hash      string
		expiresAt time.Time
		revokedAt *time.Time
	)
	err := r.pool.QueryRow(ctx, q, tokenHash).Scan(&id, &userID, &hash, &expiresAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RefreshToken{}, domain.ErrInvalidToken
	}
	if err != nil {
		return domain.RefreshToken{}, fmt.Errorf("find refresh: %w", err)
	}

	uid, err := domain.ParseUserID(userID)
	if err != nil {
		return domain.RefreshToken{}, err
	}

	return domain.RefreshToken{
		ID:        domain.RefreshTokenID(id),
		UserID:    uid,
		TokenHash: hash,
		ExpiresAt: expiresAt,
		RevokedAt: revokedAt,
	}, nil
}

// RevokeAllForUser revokes all active refresh tokens for a user.
func (r *RefreshTokenRepo) RevokeAllForUser(ctx context.Context, userID domain.UserID) error {
	const q = `
		UPDATE refresh_tokens SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, q, userID.String())
	if err != nil {
		return fmt.Errorf("revoke all refresh: %w", err)
	}
	return nil
}

// Revoke marks a refresh token revoked.
func (r *RefreshTokenRepo) Revoke(ctx context.Context, id domain.RefreshTokenID) error {
	const q = `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE id = $1 AND revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, q, string(id))
	if err != nil {
		return fmt.Errorf("revoke refresh: %w", err)
	}
	return nil
}
