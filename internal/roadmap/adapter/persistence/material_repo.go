package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// MaterialRepo implements domain.MaterialRepository.
type MaterialRepo struct {
	pool *postgres.Pool
}

// NewMaterialRepo creates MaterialRepo.
func NewMaterialRepo(pool *postgres.Pool) *MaterialRepo {
	return &MaterialRepo{pool: pool}
}

// Create inserts a material.
func (r *MaterialRepo) Create(ctx context.Context, material domain.Material) error {
	const q = `
		INSERT INTO materials (id, block_id, sort_order, title, material_type, url, required, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, q,
		string(material.ID),
		string(material.BlockID),
		material.SortOrder,
		material.Title,
		string(material.MaterialType),
		material.URL,
		material.Required,
		material.IsActive,
	)
	return err
}

// Update updates material fields.
func (r *MaterialRepo) Update(ctx context.Context, material domain.Material) error {
	const q = `
		UPDATE materials
		SET title = $2, material_type = $3, url = $4, required = $5, is_active = $6, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`
	ct, err := r.pool.Exec(ctx, q,
		string(material.ID),
		material.Title,
		string(material.MaterialType),
		material.URL,
		material.Required,
		material.IsActive,
	)
	if err != nil {
		return fmt.Errorf("update material: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrMaterialNotFound
	}
	return nil
}

// SoftDelete marks material deleted.
func (r *MaterialRepo) SoftDelete(ctx context.Context, materialID domain.MaterialID) error {
	const q = `UPDATE materials SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL`
	ct, err := r.pool.Exec(ctx, q, string(materialID))
	if err != nil {
		return fmt.Errorf("soft delete material: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrMaterialNotFound
	}
	return nil
}

// SetActive toggles is_active.
func (r *MaterialRepo) SetActive(ctx context.Context, materialID domain.MaterialID, active bool) error {
	const q = `UPDATE materials SET is_active = $2, updated_at = now() WHERE id = $1 AND deleted_at IS NULL`
	ct, err := r.pool.Exec(ctx, q, string(materialID), active)
	if err != nil {
		return fmt.Errorf("set material active: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrMaterialNotFound
	}
	return nil
}

// FindByID loads material.
func (r *MaterialRepo) FindByID(ctx context.Context, materialID domain.MaterialID) (domain.Material, error) {
	const q = `
		SELECT id, block_id, sort_order, title, material_type, url, required, is_active, created_at, updated_at, deleted_at
		FROM materials WHERE id = $1 AND deleted_at IS NULL
	`
	return scanMaterial(r.pool.QueryRow(ctx, q, string(materialID)))
}

// ListByBlock lists materials for one block.
func (r *MaterialRepo) ListByBlock(ctx context.Context, blockID domain.BlockID, studentVisibleOnly bool) ([]domain.Material, error) {
	where := `WHERE block_id = $1 AND deleted_at IS NULL`
	if studentVisibleOnly {
		where += ` AND is_active = true`
	}
	q := `
		SELECT id, block_id, sort_order, title, material_type, url, required, is_active, created_at, updated_at, deleted_at
		FROM materials ` + where + ` ORDER BY sort_order ASC`
	rows, err := r.pool.Query(ctx, q, string(blockID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMaterialRows(rows)
}

// ListByBlocks lists materials for multiple blocks.
func (r *MaterialRepo) ListByBlocks(ctx context.Context, blockIDs []domain.BlockID, studentVisibleOnly bool) ([]domain.Material, error) {
	if len(blockIDs) == 0 {
		return nil, nil
	}
	ids := make([]string, len(blockIDs))
	for i, id := range blockIDs {
		ids[i] = string(id)
	}
	where := `WHERE block_id = ANY($1) AND deleted_at IS NULL`
	if studentVisibleOnly {
		where += ` AND is_active = true`
	}
	q := `
		SELECT id, block_id, sort_order, title, material_type, url, required, is_active, created_at, updated_at, deleted_at
		FROM materials ` + where + ` ORDER BY block_id, sort_order ASC`
	rows, err := r.pool.Query(ctx, q, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMaterialRows(rows)
}

// CountActiveByBlock counts non-deleted materials.
func (r *MaterialRepo) CountActiveByBlock(ctx context.Context, blockID domain.BlockID) (int, error) {
	const q = `SELECT COUNT(*) FROM materials WHERE block_id = $1 AND deleted_at IS NULL`
	var n int
	err := r.pool.QueryRow(ctx, q, string(blockID)).Scan(&n)
	return n, err
}

// ReorderMaterials updates sort orders within a block in TX.
func (r *MaterialRepo) ReorderMaterials(ctx context.Context, blockID domain.BlockID, orders []domain.MaterialOrder) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const bump = `UPDATE materials SET sort_order = sort_order + $2, updated_at = now() WHERE block_id = $1 AND deleted_at IS NULL`
	if _, err := tx.Exec(ctx, bump, string(blockID), reorderBump); err != nil {
		return fmt.Errorf("bump material sort: %w", err)
	}

	const setOrder = `UPDATE materials SET sort_order = $3, updated_at = now() WHERE id = $1 AND block_id = $2 AND deleted_at IS NULL`
	for _, o := range orders {
		ct, err := tx.Exec(ctx, setOrder, string(o.MaterialID), string(blockID), o.SortOrder)
		if err != nil {
			return fmt.Errorf("set material sort: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return domain.ErrMaterialNotFound
		}
	}
	return tx.Commit(ctx)
}

// NextMaterialSortOrder returns next sort for block.
func (r *MaterialRepo) NextMaterialSortOrder(ctx context.Context, blockID domain.BlockID) (int, error) {
	const q = `
		SELECT COALESCE(MAX(sort_order), 0) + $2
		FROM materials
		WHERE block_id = $1 AND deleted_at IS NULL AND is_active = true
	`
	var next int
	err := r.pool.QueryRow(ctx, q, string(blockID), sortStep).Scan(&next)
	return next, err
}

func scanMaterialRows(rows pgx.Rows) ([]domain.Material, error) {
	var out []domain.Material
	for rows.Next() {
		m, err := scanMaterial(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func scanMaterial(row pgx.Row) (domain.Material, error) {
	var m domain.Material
	var id, blockID, mt string
	var deletedAt *time.Time
	err := row.Scan(
		&id,
		&blockID,
		&m.SortOrder,
		&m.Title,
		&mt,
		&m.URL,
		&m.Required,
		&m.IsActive,
		&m.CreatedAt,
		&m.UpdatedAt,
		&deletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Material{}, domain.ErrMaterialNotFound
	}
	if err != nil {
		return domain.Material{}, err
	}
	m.ID = domain.MaterialID(id)
	m.BlockID = domain.BlockID(blockID)
	m.MaterialType = domain.MaterialType(mt)
	m.DeletedAt = deletedAt
	return m, nil
}

// NewMaterialID generates material id.
func NewMaterialID() domain.MaterialID {
	return domain.MaterialID(uuid.NewString())
}
