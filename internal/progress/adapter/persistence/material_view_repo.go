package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/progress/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// MaterialViewRepo implements domain.MaterialViewRepository.
type MaterialViewRepo struct {
	pool *postgres.Pool
}

// NewMaterialViewRepo creates MaterialViewRepo.
func NewMaterialViewRepo(pool *postgres.Pool) *MaterialViewRepo {
	return &MaterialViewRepo{pool: pool}
}

// RecordFirstView inserts view or returns idempotent hit.
func (r *MaterialViewRepo) RecordFirstView(ctx context.Context, view domain.MaterialView) (bool, error) {
	id := view.ID
	if id == "" {
		id = uuid.NewString()
	}
	const q = `
		INSERT INTO material_views (id, student_id, material_id, idempotency_key)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (student_id, material_id) DO NOTHING
	`
	ct, err := querier(ctx, r.pool).Exec(ctx, q,
		id,
		string(view.StudentID),
		string(view.MaterialID),
		view.IdempotencyKey,
	)
	if err != nil {
		if isUniqueIdempotency(err) {
			return false, nil
		}
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}

// CountViewedInSet counts how many material ids were viewed.
func (r *MaterialViewRepo) CountViewedInSet(ctx context.Context, studentID domain.StudentID, materialIDs []domain.MaterialID) (int, error) {
	if len(materialIDs) == 0 {
		return 0, nil
	}
	ids := make([]string, len(materialIDs))
	for i, id := range materialIDs {
		ids[i] = string(id)
	}
	const q = `
		SELECT COUNT(*) FROM material_views
		WHERE student_id = $1 AND material_id = ANY($2)
	`
	var n int
	err := querier(ctx, r.pool).QueryRow(ctx, q, string(studentID), ids).Scan(&n)
	return n, err
}

// ListViewedMaterialIDs lists viewed materials in a block.
func (r *MaterialViewRepo) ListViewedMaterialIDs(ctx context.Context, studentID domain.StudentID, blockID domain.BlockID) ([]domain.MaterialID, error) {
	const q = `
		SELECT v.material_id
		FROM material_views v
		INNER JOIN materials m ON m.id = v.material_id
		WHERE v.student_id = $1 AND m.block_id = $2 AND m.deleted_at IS NULL
	`
	rows, err := r.pool.Query(ctx, q, string(studentID), string(blockID))
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

// FirstViewedAtByMaterials returns RFC3339 timestamps keyed by material id.
func (r *MaterialViewRepo) FirstViewedAtByMaterials(ctx context.Context, studentID domain.StudentID, materialIDs []domain.MaterialID) (map[domain.MaterialID]string, error) {
	out := make(map[domain.MaterialID]string)
	if len(materialIDs) == 0 {
		return out, nil
	}
	ids := make([]string, len(materialIDs))
	for i, id := range materialIDs {
		ids[i] = string(id)
	}
	const q = `
		SELECT material_id, first_viewed_at
		FROM material_views
		WHERE student_id = $1 AND material_id = ANY($2)
	`
	rows, err := r.pool.Query(ctx, q, string(studentID), ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var mid string
		var ts time.Time
		if err := rows.Scan(&mid, &ts); err != nil {
			return nil, err
		}
		out[domain.MaterialID(mid)] = ts.UTC().Format(time.RFC3339)
	}
	return out, rows.Err()
}

func isUniqueIdempotency(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
