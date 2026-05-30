package persistence

import (
	"context"
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/user/domain"
)

// ProfileBootstrap ensures a profile row exists (avoids import cycle with profile module).
type ProfileBootstrap struct {
	pool *postgres.Pool
}

// NewProfileBootstrap creates ProfileBootstrap.
func NewProfileBootstrap(pool *postgres.Pool) *ProfileBootstrap {
	return &ProfileBootstrap{pool: pool}
}

// EnsureEmpty implements domain.ProfileBootstrap.
func (b *ProfileBootstrap) EnsureEmpty(ctx context.Context, userID domain.UserID) error {
	const q = `INSERT INTO profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING`
	_, err := b.pool.Exec(ctx, q, string(userID))
	if err != nil {
		return fmt.Errorf("ensure profile: %w", err)
	}
	return nil
}
