package persistence

import (
	"context"
	"errors"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/finalcheck/domain"
	"github.com/jackc/pgx/v5"
)

// BuddyScopeReader implements BuddyScopePort.
type BuddyScopeReader struct {
	pool *postgres.Pool
}

// NewBuddyScopeReader creates BuddyScopeReader.
func NewBuddyScopeReader(pool *postgres.Pool) *BuddyScopeReader {
	return &BuddyScopeReader{pool: pool}
}

// IsActiveBuddyOf checks assignment.
func (r *BuddyScopeReader) IsActiveBuddyOf(ctx context.Context, buddyID domain.UserID, studentID domain.StudentID) (bool, error) {
	const q = `
		SELECT 1 FROM buddy_assignments
		WHERE buddy_id = $1 AND student_id = $2 AND active = true AND deleted_at IS NULL
	`
	var one int
	err := r.pool.QueryRow(ctx, q, string(buddyID), string(studentID)).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
