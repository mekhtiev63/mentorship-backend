package persistence

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
)

// RoadmapStatsReader implements RoadmapStatsPort.
type RoadmapStatsReader struct {
	pool *postgres.Pool
}

// NewRoadmapStatsReader creates RoadmapStatsReader.
func NewRoadmapStatsReader(pool *postgres.Pool) *RoadmapStatsReader {
	return &RoadmapStatsReader{pool: pool}
}

// CountPublishedBlocks counts student-visible published blocks.
func (r *RoadmapStatsReader) CountPublishedBlocks(ctx context.Context) (int, error) {
	const q = `
		SELECT COUNT(*) FROM roadmap_blocks
		WHERE deleted_at IS NULL AND is_active = true AND status = 'published'
	`
	var n int
	err := r.pool.QueryRow(ctx, q).Scan(&n)
	return n, err
}
