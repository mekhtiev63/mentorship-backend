package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/achievement/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// UserAchievementRepo implements UserAchievementRepository.
type UserAchievementRepo struct {
	pool *postgres.Pool
}

// NewUserAchievementRepo creates UserAchievementRepo.
func NewUserAchievementRepo(pool *postgres.Pool) *UserAchievementRepo {
	return &UserAchievementRepo{pool: pool}
}

// Exists reports whether user already has achievement.
func (r *UserAchievementRepo) Exists(ctx context.Context, userID domain.UserID, code domain.AchievementCode) (bool, error) {
	const q = `SELECT 1 FROM user_achievements WHERE user_id = $1 AND achievement_code = $2`
	var one int
	err := db(ctx, r.pool).QueryRow(ctx, q, string(userID), string(code)).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// Grant inserts grant idempotently.
func (r *UserAchievementRepo) Grant(ctx context.Context, a domain.UserAchievement) (bool, error) {
	const q = `
		INSERT INTO user_achievements (user_id, achievement_code, granted_at, source_event_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, achievement_code) DO NOTHING
	`
	ct, err := db(ctx, r.pool).Exec(ctx, q,
		string(a.UserID),
		string(a.AchievementCode),
		a.GrantedAt,
		string(a.SourceEventID),
	)
	if err != nil {
		if isSourceConflict(err) {
			return false, nil
		}
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}

// ListByUser lists grants for user.
func (r *UserAchievementRepo) ListByUser(ctx context.Context, userID domain.UserID) ([]domain.UserAchievement, error) {
	const q = `
		SELECT achievement_code, granted_at, source_event_id
		FROM user_achievements
		WHERE user_id = $1
		ORDER BY granted_at DESC
	`
	rows, err := r.pool.Query(ctx, q, string(userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.UserAchievement
	for rows.Next() {
		var a domain.UserAchievement
		a.UserID = userID
		var code string
		var source string
		if err := rows.Scan(&code, &a.GrantedAt, &source); err != nil {
			return nil, err
		}
		a.AchievementCode = domain.AchievementCode(code)
		a.SourceEventID = domain.SourceEventID(source)
		out = append(out, a)
	}
	return out, rows.Err()
}

func isSourceConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// NowUTC helper for grants.
func NowUTC() time.Time { return time.Now().UTC() }
