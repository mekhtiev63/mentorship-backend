package persistence

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/finalcheck/domain"
)

// RoadmapCompletionReader implements RoadmapCompletionPort.
type RoadmapCompletionReader struct {
	pool *postgres.Pool
}

// NewRoadmapCompletionReader creates RoadmapCompletionReader.
func NewRoadmapCompletionReader(pool *postgres.Pool) *RoadmapCompletionReader {
	return &RoadmapCompletionReader{pool: pool}
}

// IsProgramCompleted reports whether all published blocks are approved for student.
func (r *RoadmapCompletionReader) IsProgramCompleted(ctx context.Context, studentID domain.StudentID) (bool, error) {
	const publishedQ = `
		SELECT COUNT(*) FROM roadmap_blocks
		WHERE deleted_at IS NULL AND is_active = true AND status = 'published'
	`
	var published int
	if err := r.pool.QueryRow(ctx, publishedQ).Scan(&published); err != nil {
		return false, err
	}
	if published == 0 {
		return false, nil
	}
	const approvedQ = `
		SELECT COUNT(*) FROM student_block_progress
		WHERE student_id = $1 AND status = 'approved'
	`
	var approved int
	if err := r.pool.QueryRow(ctx, approvedQ, string(studentID)).Scan(&approved); err != nil {
		return false, err
	}
	return approved >= published, nil
}
