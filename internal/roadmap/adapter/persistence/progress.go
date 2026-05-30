package persistence

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
	"github.com/jackc/pgx/v5"
)

// ProgressReader checks student_block_progress.
type ProgressReader struct {
	pool *postgres.Pool
}

// NewProgressReader creates ProgressReader.
func NewProgressReader(pool *postgres.Pool) *ProgressReader {
	return &ProgressReader{pool: pool}
}

// HasProgressForBlock reports whether any progress row exists for the block.
func (r *ProgressReader) HasProgressForBlock(ctx context.Context, blockID domain.BlockID) (bool, error) {
	const q = `SELECT 1 FROM student_block_progress WHERE block_id = $1 LIMIT 1`
	var one int
	err := r.pool.QueryRow(ctx, q, string(blockID)).Scan(&one)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
