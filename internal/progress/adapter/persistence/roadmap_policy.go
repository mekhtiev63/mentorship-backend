package persistence

import (
	"context"
	"errors"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/progress/domain"
	"github.com/jackc/pgx/v5"
)

// RoadmapPolicyReader implements domain.RoadmapProgressPolicyPort.
type RoadmapPolicyReader struct {
	pool *postgres.Pool
}

// NewRoadmapPolicyReader creates RoadmapPolicyReader.
func NewRoadmapPolicyReader(pool *postgres.Pool) *RoadmapPolicyReader {
	return &RoadmapPolicyReader{pool: pool}
}

// MaterialBlockID resolves block for material.
func (r *RoadmapPolicyReader) MaterialBlockID(ctx context.Context, materialID domain.MaterialID) (domain.BlockID, error) {
	const q = `SELECT block_id FROM materials WHERE id = $1 AND deleted_at IS NULL`
	var blockID string
	err := r.pool.QueryRow(ctx, q, string(materialID)).Scan(&blockID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrMaterialNotFound
	}
	return domain.BlockID(blockID), err
}

// IsMaterialVisibleToStudent checks student-visible material.
func (r *RoadmapPolicyReader) IsMaterialVisibleToStudent(ctx context.Context, materialID domain.MaterialID) (bool, error) {
	const q = `
		SELECT 1
		FROM materials m
		INNER JOIN roadmap_blocks b ON b.id = m.block_id
		WHERE m.id = $1
		  AND m.deleted_at IS NULL AND m.is_active = true
		  AND b.deleted_at IS NULL AND b.is_active = true AND b.status = 'published'
	`
	var one int
	err := r.pool.QueryRow(ctx, q, string(materialID)).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// IsBlockVisibleToStudent checks published active block.
func (r *RoadmapPolicyReader) IsBlockVisibleToStudent(ctx context.Context, blockID domain.BlockID) (bool, error) {
	const q = `
		SELECT 1 FROM roadmap_blocks
		WHERE id = $1 AND deleted_at IS NULL AND is_active = true AND status = 'published'
	`
	var one int
	err := r.pool.QueryRow(ctx, q, string(blockID)).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// ListRequiredMaterialIDs returns required visible materials.
func (r *RoadmapPolicyReader) ListRequiredMaterialIDs(ctx context.Context, blockID domain.BlockID) ([]domain.MaterialID, error) {
	const q = `
		SELECT id FROM materials
		WHERE block_id = $1 AND deleted_at IS NULL AND is_active = true AND required = true
		ORDER BY sort_order ASC
	`
	rows, err := r.pool.Query(ctx, q, string(blockID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.MaterialID
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, domain.MaterialID(id))
	}
	return out, rows.Err()
}

// ListPublishedBlocksOrdered returns published blocks in order.
func (r *RoadmapPolicyReader) ListPublishedBlocksOrdered(ctx context.Context) ([]domain.RoadmapBlockRef, error) {
	const q = `
		SELECT id, sort_order, title
		FROM roadmap_blocks
		WHERE deleted_at IS NULL AND is_active = true AND status = 'published'
		ORDER BY sort_order ASC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RoadmapBlockRef
	for rows.Next() {
		var ref domain.RoadmapBlockRef
		var id string
		if err := rows.Scan(&id, &ref.SortOrder, &ref.Title); err != nil {
			return nil, err
		}
		ref.BlockID = domain.BlockID(id)
		out = append(out, ref)
	}
	return out, rows.Err()
}
