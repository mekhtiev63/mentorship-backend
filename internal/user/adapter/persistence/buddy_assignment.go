package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/user/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/jackc/pgx/v5"
)

// BuddyAssignmentRepo implements domain.BuddyAssignmentRepository.
type BuddyAssignmentRepo struct {
	pool *postgres.Pool
}

// NewBuddyAssignmentRepo creates BuddyAssignmentRepo.
func NewBuddyAssignmentRepo(pool *postgres.Pool) *BuddyAssignmentRepo {
	return &BuddyAssignmentRepo{pool: pool}
}

// Assign deactivates previous active assignment and creates a new one.
func (r *BuddyAssignmentRepo) Assign(ctx context.Context, studentID, buddyID domain.UserID) (domain.BuddyAssignment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.BuddyAssignment{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		UPDATE buddy_assignments
		SET active = false, updated_at = now()
		WHERE student_id = $1 AND active = true AND deleted_at IS NULL
	`, string(studentID))
	if err != nil {
		return domain.BuddyAssignment{}, err
	}

	const ins = `
		INSERT INTO buddy_assignments (student_id, buddy_id, active)
		VALUES ($1, $2, true)
		RETURNING id, student_id, buddy_id, active, valid_from, valid_to, created_at, updated_at, deleted_at
	`
	row := tx.QueryRow(ctx, ins, string(studentID), string(buddyID))
	a, err := scanAssignment(row)
	if err != nil {
		return domain.BuddyAssignment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.BuddyAssignment{}, err
	}
	return a, nil
}

// DeactivateActiveForStudent deactivates active assignment.
func (r *BuddyAssignmentRepo) DeactivateActiveForStudent(ctx context.Context, studentID domain.UserID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE buddy_assignments SET active = false, updated_at = now()
		WHERE student_id = $1 AND active = true AND deleted_at IS NULL
	`, string(studentID))
	return err
}

// FindActiveByStudent returns active assignment.
func (r *BuddyAssignmentRepo) FindActiveByStudent(ctx context.Context, studentID domain.UserID) (domain.BuddyAssignment, error) {
	const q = `
		SELECT id, student_id, buddy_id, active, valid_from, valid_to, created_at, updated_at, deleted_at
		FROM buddy_assignments
		WHERE student_id = $1 AND active = true AND deleted_at IS NULL
	`
	return scanAssignment(r.pool.QueryRow(ctx, q, string(studentID)))
}

// ListActiveStudentsForBuddy lists students for buddy.
func (r *BuddyAssignmentRepo) ListActiveStudentsForBuddy(ctx context.Context, buddyID domain.UserID, page pagination.Params) ([]domain.StudentSummary, int64, error) {
	const countQ = `
		SELECT COUNT(*)
		FROM buddy_assignments ba
		JOIN users u ON u.id = ba.student_id
		WHERE ba.buddy_id = $1 AND ba.active = true AND ba.deleted_at IS NULL AND u.deleted_at IS NULL
	`
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, string(buddyID)).Scan(&total); err != nil {
		return nil, 0, err
	}

	const q = `
		SELECT u.id, u.email
		FROM buddy_assignments ba
		JOIN users u ON u.id = ba.student_id
		WHERE ba.buddy_id = $1 AND ba.active = true AND ba.deleted_at IS NULL AND u.deleted_at IS NULL
		ORDER BY u.email
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, q, string(buddyID), page.PageSize, page.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []domain.StudentSummary
	for rows.Next() {
		var id, email string
		if err := rows.Scan(&id, &email); err != nil {
			return nil, 0, err
		}
		out = append(out, domain.StudentSummary{ID: domain.UserID(id), Email: domain.Email(email)})
	}
	return out, total, rows.Err()
}

// SoftDelete soft-deletes assignment and deactivates.
func (r *BuddyAssignmentRepo) SoftDelete(ctx context.Context, id domain.AssignmentID) error {
	const q = `
		UPDATE buddy_assignments
		SET deleted_at = now(), active = false, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`
	ct, err := r.pool.Exec(ctx, q, string(id))
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrAssignmentNotFound
	}
	return nil
}

// FindByID loads assignment by id.
func (r *BuddyAssignmentRepo) FindByID(ctx context.Context, id domain.AssignmentID) (domain.BuddyAssignment, error) {
	const q = `
		SELECT id, student_id, buddy_id, active, valid_from, valid_to, created_at, updated_at, deleted_at
		FROM buddy_assignments WHERE id = $1 AND deleted_at IS NULL
	`
	return scanAssignment(r.pool.QueryRow(ctx, q, string(id)))
}

// DeactivateForBuddy deactivates all active assignments for a buddy.
func (r *BuddyAssignmentRepo) DeactivateForBuddy(ctx context.Context, buddyID domain.UserID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE buddy_assignments SET active = false, updated_at = now()
		WHERE buddy_id = $1 AND active = true AND deleted_at IS NULL
	`, string(buddyID))
	return err
}

// IsAssignedBuddy checks active assignment.
func (r *BuddyAssignmentRepo) IsAssignedBuddy(ctx context.Context, buddyID, studentID domain.UserID) (bool, error) {
	const q = `
		SELECT 1 FROM buddy_assignments
		WHERE buddy_id = $1 AND student_id = $2 AND active = true AND deleted_at IS NULL
	`
	var one int
	err := r.pool.QueryRow(ctx, q, string(buddyID), string(studentID)).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func scanAssignment(row pgx.Row) (domain.BuddyAssignment, error) {
	var (
		id, studentID, buddyID string
		active                 bool
		validFrom              time.Time
		validTo                *time.Time
		createdAt, updatedAt   time.Time
		deletedAt              *time.Time
	)
	err := row.Scan(&id, &studentID, &buddyID, &active, &validFrom, &validTo, &createdAt, &updatedAt, &deletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.BuddyAssignment{}, domain.ErrAssignmentNotFound
	}
	if err != nil {
		return domain.BuddyAssignment{}, err
	}
	return domain.BuddyAssignment{
		ID:        domain.AssignmentID(id),
		StudentID: domain.UserID(studentID),
		BuddyID:   domain.UserID(buddyID),
		Active:    active,
		ValidFrom: validFrom,
		ValidTo:   validTo,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	}, nil
}
