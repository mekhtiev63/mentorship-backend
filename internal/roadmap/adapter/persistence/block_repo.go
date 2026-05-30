package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// BlockRepo implements domain.BlockRepository.
type BlockRepo struct {
	pool *postgres.Pool
}

// NewBlockRepo creates BlockRepo.
func NewBlockRepo(pool *postgres.Pool) *BlockRepo {
	return &BlockRepo{pool: pool}
}

// Create inserts a roadmap block.
func (r *BlockRepo) Create(ctx context.Context, block domain.RoadmapBlock) error {
	skills := block.ExpectedSkills
	if skills == nil {
		skills = []string{}
	}
	const q = `
		INSERT INTO roadmap_blocks (id, sort_order, title, description, expected_skills, status, is_active, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, q,
		string(block.ID),
		block.SortOrder,
		block.Title,
		block.Description,
		skills,
		string(block.Status),
		block.IsActive,
		block.PublishedAt,
	)
	return err
}

// Update updates block fields (not sort_order).
func (r *BlockRepo) Update(ctx context.Context, block domain.RoadmapBlock) error {
	const q = `
		UPDATE roadmap_blocks
		SET title = $2, description = $3, status = $4, is_active = $5, published_at = $6, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`
	ct, err := r.pool.Exec(ctx, q,
		string(block.ID),
		block.Title,
		block.Description,
		string(block.Status),
		block.IsActive,
		block.PublishedAt,
	)
	if err != nil {
		return fmt.Errorf("update block: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SoftDelete marks block deleted.
func (r *BlockRepo) SoftDelete(ctx context.Context, blockID domain.BlockID) error {
	const q = `UPDATE roadmap_blocks SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL`
	ct, err := r.pool.Exec(ctx, q, string(blockID))
	if err != nil {
		return fmt.Errorf("soft delete block: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SoftDeleteMaterialsByBlock soft-deletes all materials in a block.
func (r *BlockRepo) SoftDeleteMaterialsByBlock(ctx context.Context, blockID domain.BlockID) error {
	const q = `UPDATE materials SET deleted_at = now(), updated_at = now() WHERE block_id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, q, string(blockID))
	return err
}

// SetActive toggles is_active.
func (r *BlockRepo) SetActive(ctx context.Context, blockID domain.BlockID, active bool) error {
	const q = `UPDATE roadmap_blocks SET is_active = $2, updated_at = now() WHERE id = $1 AND deleted_at IS NULL`
	ct, err := r.pool.Exec(ctx, q, string(blockID), active)
	if err != nil {
		return fmt.Errorf("set block active: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// FindByID loads block including soft-deleted for admin? Plan: admin can see deleted - use separate. For now admin only non-deleted.
func (r *BlockRepo) FindByID(ctx context.Context, blockID domain.BlockID) (domain.RoadmapBlock, error) {
	const q = `
		SELECT id, sort_order, title, description, expected_skills, status, is_active, published_at, created_at, updated_at, deleted_at
		FROM roadmap_blocks WHERE id = $1 AND deleted_at IS NULL
	`
	return scanBlock(r.pool.QueryRow(ctx, q, string(blockID)))
}

// FindByIDIncludingDeleted loads block for admin edge cases - skip for MVP

// ListAdmin lists blocks for admin.
func (r *BlockRepo) ListAdmin(ctx context.Context, filter domain.AdminBlockFilter, page pagination.Params) ([]domain.RoadmapBlock, int64, error) {
	where := `WHERE deleted_at IS NULL`
	args := []any{}
	n := 1
	if filter.Status != nil {
		where += fmt.Sprintf(` AND status = $%d`, n)
		args = append(args, string(*filter.Status))
		n++
	}
	if filter.IsActive != nil {
		where += fmt.Sprintf(` AND is_active = $%d`, n)
		args = append(args, *filter.IsActive)
		n++
	}

	var total int64
	countQ := `SELECT COUNT(*) FROM roadmap_blocks ` + where
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, page.PageSize, page.Offset)
	listQ := `
		SELECT id, sort_order, title, description, expected_skills, status, is_active, published_at, created_at, updated_at, deleted_at
		FROM roadmap_blocks ` + where + `
		ORDER BY sort_order ASC
		LIMIT $` + fmt.Sprint(n) + ` OFFSET $` + fmt.Sprint(n+1)
	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var blocks []domain.RoadmapBlock
	for rows.Next() {
		b, err := scanBlock(rows)
		if err != nil {
			return nil, 0, err
		}
		blocks = append(blocks, b)
	}
	return blocks, total, rows.Err()
}

// ListPublishedBlocks returns student-visible blocks.
func (r *BlockRepo) ListPublishedBlocks(ctx context.Context) ([]domain.RoadmapBlock, error) {
	const q = `
		SELECT id, sort_order, title, description, expected_skills, status, is_active, published_at, created_at, updated_at, deleted_at
		FROM roadmap_blocks
		WHERE deleted_at IS NULL AND is_active = true AND status = 'published'
		ORDER BY sort_order ASC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var blocks []domain.RoadmapBlock
	for rows.Next() {
		b, err := scanBlock(rows)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, rows.Err()
}

// ReorderBlocks updates sort orders in a transaction.
func (r *BlockRepo) ReorderBlocks(ctx context.Context, orders []domain.BlockOrder) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const bump = `UPDATE roadmap_blocks SET sort_order = sort_order + $1, updated_at = now() WHERE deleted_at IS NULL`
	if _, err := tx.Exec(ctx, bump, reorderBump); err != nil {
		return fmt.Errorf("bump block sort: %w", err)
	}

	const setOrder = `UPDATE roadmap_blocks SET sort_order = $2, updated_at = now() WHERE id = $1 AND deleted_at IS NULL`
	for _, o := range orders {
		ct, err := tx.Exec(ctx, setOrder, string(o.BlockID), o.SortOrder)
		if err != nil {
			return fmt.Errorf("set block sort: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return domain.ErrNotFound
		}
	}
	return tx.Commit(ctx)
}

// NextBlockSortOrder returns next sort slot for new active blocks.
func (r *BlockRepo) NextBlockSortOrder(ctx context.Context) (int, error) {
	const q = `
		SELECT COALESCE(MAX(sort_order), 0) + $1
		FROM roadmap_blocks
		WHERE deleted_at IS NULL AND is_active = true
	`
	var next int
	err := r.pool.QueryRow(ctx, q, sortStep).Scan(&next)
	return next, err
}

func scanBlock(row pgx.Row) (domain.RoadmapBlock, error) {
	var b domain.RoadmapBlock
	var id string
	var status string
	var publishedAt *time.Time
	var deletedAt *time.Time
	err := row.Scan(
		&id,
		&b.SortOrder,
		&b.Title,
		&b.Description,
		&b.ExpectedSkills,
		&status,
		&b.IsActive,
		&publishedAt,
		&b.CreatedAt,
		&b.UpdatedAt,
		&deletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RoadmapBlock{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.RoadmapBlock{}, err
	}
	b.ID = domain.BlockID(id)
	b.Status = domain.BlockStatus(status)
	b.PublishedAt = publishedAt
	b.DeletedAt = deletedAt
	return b, nil
}

// NewBlockID generates a block id.
func NewBlockID() domain.BlockID {
	return domain.BlockID(uuid.NewString())
}
