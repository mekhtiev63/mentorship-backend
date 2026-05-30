package persistence

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/progress/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/jackc/pgx/v5"
)

// BuddyScopeReader implements domain.BuddyScopePort.
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
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// ListActiveStudentIDsForBuddy lists student ids.
func (r *BuddyScopeReader) ListActiveStudentIDsForBuddy(ctx context.Context, buddyID domain.UserID, page pagination.Params) ([]domain.StudentID, int64, error) {
	const countQ = `
		SELECT COUNT(*) FROM buddy_assignments
		WHERE buddy_id = $1 AND active = true AND deleted_at IS NULL
	`
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, string(buddyID)).Scan(&total); err != nil {
		return nil, 0, err
	}
	const q = `
		SELECT student_id FROM buddy_assignments
		WHERE buddy_id = $1 AND active = true AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, q, string(buddyID), page.PageSize, page.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.StudentID
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		out = append(out, domain.StudentID(id))
	}
	return out, total, rows.Err()
}
