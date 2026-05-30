package persistence

import (
	"context"
	"errors"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/oneonone/domain"
	"github.com/jackc/pgx/v5"
)

// BuddyReader resolves active buddy assignments.
type BuddyReader struct {
	pool *postgres.Pool
}

// NewBuddyReader creates BuddyReader.
func NewBuddyReader(pool *postgres.Pool) *BuddyReader {
	return &BuddyReader{pool: pool}
}

// GetActiveBuddyID returns buddy for student.
func (r *BuddyReader) GetActiveBuddyID(ctx context.Context, studentID domain.StudentID) (domain.BuddyID, error) {
	const q = `
		SELECT buddy_id
		FROM buddy_assignments
		WHERE student_id = $1 AND active = true AND deleted_at IS NULL
	`
	var buddyID string
	err := r.pool.QueryRow(ctx, q, string(studentID)).Scan(&buddyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrForbidden
	}
	if err != nil {
		return "", err
	}
	return domain.BuddyID(buddyID), nil
}
