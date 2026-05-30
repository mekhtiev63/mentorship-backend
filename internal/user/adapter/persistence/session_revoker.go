package persistence

import (
	"context"
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/user/domain"
)

// SessionRevoker revokes refresh tokens when a user is soft-deleted.
type SessionRevoker struct {
	pool *postgres.Pool
}

// NewSessionRevoker creates SessionRevoker.
func NewSessionRevoker(pool *postgres.Pool) *SessionRevoker {
	return &SessionRevoker{pool: pool}
}

// RevokeAllForUser implements domain.SessionRevoker.
func (s *SessionRevoker) RevokeAllForUser(ctx context.Context, userID domain.UserID) error {
	const q = `UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`
	_, err := s.pool.Exec(ctx, q, string(userID))
	if err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}
	return nil
}
