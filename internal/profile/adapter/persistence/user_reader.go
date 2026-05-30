package persistence

import (
	"context"
	"errors"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/profile/domain"
	userpersistence "github.com/go-mentorship-platform/backend/internal/user/adapter/persistence"
	userdomain "github.com/go-mentorship-platform/backend/internal/user/domain"
	"github.com/jackc/pgx/v5"
)

// UserExistsRepo checks users table.
type UserExistsRepo struct {
	pool *postgres.Pool
}

// NewUserExistsRepo creates UserExistsRepo.
func NewUserExistsRepo(pool *postgres.Pool) *UserExistsRepo {
	return &UserExistsRepo{pool: pool}
}

// ExistsActive implements domain.UserExistsReader.
func (r *UserExistsRepo) ExistsActive(ctx context.Context, userID domain.UserID) (bool, error) {
	const q = `SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL`
	var one int
	err := r.pool.QueryRow(ctx, q, string(userID)).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// BuddyReader adapts user buddy repository.
type BuddyReader struct {
	buddy *userpersistence.BuddyAssignmentRepo
}

// NewBuddyReader creates BuddyReader.
func NewBuddyReader(buddy *userpersistence.BuddyAssignmentRepo) *BuddyReader {
	return &BuddyReader{buddy: buddy}
}

// IsAssignedBuddy implements domain.BuddyAssignmentReader.
func (r *BuddyReader) IsAssignedBuddy(ctx context.Context, buddyID, studentID domain.UserID) (bool, error) {
	return r.buddy.IsAssignedBuddy(ctx, userdomain.UserID(buddyID), userdomain.UserID(studentID))
}
