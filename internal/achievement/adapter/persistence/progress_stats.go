package persistence

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/achievement/domain"
)

// ProgressStatsReader implements ProgressStatsPort.
type ProgressStatsReader struct {
	pool *postgres.Pool
}

// NewProgressStatsReader creates ProgressStatsReader.
func NewProgressStatsReader(pool *postgres.Pool) *ProgressStatsReader {
	return &ProgressStatsReader{pool: pool}
}

// CountApprovedBlocks counts approved block progress rows.
func (r *ProgressStatsReader) CountApprovedBlocks(ctx context.Context, studentID domain.UserID) (int, error) {
	const q = `
		SELECT COUNT(*) FROM student_block_progress
		WHERE student_id = $1 AND status = 'approved'
	`
	var n int
	err := r.pool.QueryRow(ctx, q, string(studentID)).Scan(&n)
	return n, err
}

// CountMaterialViews counts material views for student.
func (r *ProgressStatsReader) CountMaterialViews(ctx context.Context, studentID domain.UserID) (int, error) {
	const q = `SELECT COUNT(*) FROM material_views WHERE student_id = $1`
	var n int
	err := r.pool.QueryRow(ctx, q, string(studentID)).Scan(&n)
	return n, err
}
