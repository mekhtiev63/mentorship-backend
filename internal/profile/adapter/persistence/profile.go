package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/profile/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ProfileRepo implements domain.ProfileRepository.
type ProfileRepo struct {
	pool *postgres.Pool
}

// NewProfileRepo creates ProfileRepo.
func NewProfileRepo(pool *postgres.Pool) *ProfileRepo {
	return &ProfileRepo{pool: pool}
}

// EnsureEmpty creates an empty profile if missing.
func (r *ProfileRepo) EnsureEmpty(ctx context.Context, userID domain.UserID) error {
	const q = `
		INSERT INTO profiles (user_id) VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, q, string(userID))
	return err
}

// GetByUserID loads profile.
func (r *ProfileRepo) GetByUserID(ctx context.Context, userID domain.UserID) (domain.Profile, error) {
	const q = `
		SELECT user_id, display_name, bio, avatar_url, telegram_username, visibility::text, created_at, updated_at
		FROM profiles WHERE user_id = $1
	`
	return scanProfile(r.pool.QueryRow(ctx, q, string(userID)))
}

// Update updates profile fields.
func (r *ProfileRepo) Update(ctx context.Context, profile domain.Profile) error {
	const q = `
		UPDATE profiles
		SET display_name = $2, bio = $3, avatar_url = $4, telegram_username = $5,
		    visibility = $6::profile_visibility, updated_at = now()
		WHERE user_id = $1
	`
	_, err := r.pool.Exec(ctx, q,
		string(profile.UserID),
		profile.DisplayName,
		profile.Bio,
		profile.AvatarURL,
		profile.TelegramUsername,
		string(profile.Visibility),
	)
	if isUniqueViolation(err) {
		return domain.ErrTelegramTaken
	}
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

func scanProfile(row pgx.Row) (domain.Profile, error) {
	var (
		userID, displayName, bio, visibility string
		avatarURL, telegram                  *string
		createdAt, updatedAt                 time.Time
	)
	err := row.Scan(&userID, &displayName, &bio, &avatarURL, &telegram, &visibility, &createdAt, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Profile{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Profile{}, err
	}
	return domain.Profile{
		UserID:           domain.UserID(userID),
		DisplayName:      displayName,
		Bio:              bio,
		AvatarURL:        avatarURL,
		TelegramUsername: telegram,
		Visibility:       domain.Visibility(visibility),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
