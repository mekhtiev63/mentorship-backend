package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/identity/domain"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
)

// PreferencesRepo stores user_preferences.
type PreferencesRepo struct {
	pool *postgres.Pool
}

// NewPreferencesRepo creates a PreferencesRepo.
func NewPreferencesRepo(pool *postgres.Pool) *PreferencesRepo {
	return &PreferencesRepo{pool: pool}
}

// Get returns preferences or empty active role when missing.
func (r *PreferencesRepo) Get(ctx context.Context, userID domain.UserID) (domain.UserPreferences, error) {
	const q = `
		SELECT active_role::text
		FROM user_preferences
		WHERE user_id = $1
	`
	var roleText *string
	err := r.pool.QueryRow(ctx, q, userID.String()).Scan(&roleText)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.UserPreferences{UserID: userID}, nil
	}
	if err != nil {
		return domain.UserPreferences{}, fmt.Errorf("get preferences: %w", err)
	}

	prefs := domain.UserPreferences{UserID: userID}
	if roleText != nil {
		role := domain.Role(*roleText)
		prefs.ActiveRole = &role
	}
	return prefs, nil
}

// SetActiveRole upserts the active role (NULL clears selection).
func (r *PreferencesRepo) SetActiveRole(ctx context.Context, userID domain.UserID, role *domain.Role) error {
	const q = `
		INSERT INTO user_preferences (user_id, active_role, updated_at)
		VALUES ($1, $2::app_role, now())
		ON CONFLICT (user_id) DO UPDATE
		SET active_role = EXCLUDED.active_role,
		    updated_at = now()
	`
	var roleArg any
	if role != nil {
		roleArg = string(*role)
	}
	_, err := r.pool.Exec(ctx, q, userID.String(), roleArg)
	if err != nil {
		return fmt.Errorf("set active role: %w", err)
	}
	return nil
}
