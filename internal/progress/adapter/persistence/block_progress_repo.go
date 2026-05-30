package persistence

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/progress/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/jackc/pgx/v5"
)

// BlockProgressRepo implements domain.BlockProgressRepository.
type BlockProgressRepo struct {
	pool *postgres.Pool
}

// NewBlockProgressRepo creates BlockProgressRepo.
func NewBlockProgressRepo(pool *postgres.Pool) *BlockProgressRepo {
	return &BlockProgressRepo{pool: pool}
}

// Get loads progress or returns synthetic not_started.
func (r *BlockProgressRepo) Get(ctx context.Context, key domain.BlockProgressKey) (domain.BlockProgress, error) {
	const q = `
		SELECT status, submitted_at, approved_by, approved_at, rejected_at, reject_reason, created_at, updated_at
		FROM student_block_progress
		WHERE student_id = $1 AND block_id = $2
	`
	p := domain.BlockProgress{
		StudentID: key.StudentID,
		BlockID:   key.BlockID,
		Status:    domain.StatusNotStarted,
	}
	var status string
	var approvedBy *string
	err := querier(ctx, r.pool).QueryRow(ctx, q, string(key.StudentID), string(key.BlockID)).Scan(
		&status,
		&p.SubmittedAt,
		&approvedBy,
		&p.ApprovedAt,
		&p.RejectedAt,
		&p.RejectReason,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, nil
	}
	if err != nil {
		return domain.BlockProgress{}, err
	}
	p.Status = domain.ProgressStatus(status)
	p.Exists = true
	if approvedBy != nil {
		u := domain.UserID(*approvedBy)
		p.ApprovedBy = &u
	}
	return p, nil
}

// Insert creates a new progress row.
func (r *BlockProgressRepo) Insert(ctx context.Context, progress domain.BlockProgress) error {
	const q = `
		INSERT INTO student_block_progress (student_id, block_id, status, submitted_at, approved_by, approved_at, rejected_at, reject_reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	var approvedBy *string
	if progress.ApprovedBy != nil {
		s := string(*progress.ApprovedBy)
		approvedBy = &s
	}
	_, err := querier(ctx, r.pool).Exec(ctx, q,
		string(progress.StudentID),
		string(progress.BlockID),
		string(progress.Status),
		progress.SubmittedAt,
		approvedBy,
		progress.ApprovedAt,
		progress.RejectedAt,
		progress.RejectReason,
	)
	return err
}

// Save updates progress with optimistic status check.
func (r *BlockProgressRepo) Save(ctx context.Context, progress domain.BlockProgress, expectedStatus domain.ProgressStatus) error {
	const q = `
		UPDATE student_block_progress
		SET status = $3, submitted_at = $4, approved_by = $5, approved_at = $6,
		    rejected_at = $7, reject_reason = $8, updated_at = now()
		WHERE student_id = $1 AND block_id = $2 AND status = $9
	`
	var approvedBy *string
	if progress.ApprovedBy != nil {
		s := string(*progress.ApprovedBy)
		approvedBy = &s
	}
	ct, err := querier(ctx, r.pool).Exec(ctx, q,
		string(progress.StudentID),
		string(progress.BlockID),
		string(progress.Status),
		progress.SubmittedAt,
		approvedBy,
		progress.ApprovedAt,
		progress.RejectedAt,
		progress.RejectReason,
		string(expectedStatus),
	)
	if err != nil {
		return fmt.Errorf("save progress: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

// ListByStudent returns all progress rows for a student.
func (r *BlockProgressRepo) ListByStudent(ctx context.Context, studentID domain.StudentID) ([]domain.BlockProgress, error) {
	const q = `
		SELECT block_id, status, submitted_at, approved_by, approved_at, rejected_at, reject_reason, created_at, updated_at
		FROM student_block_progress
		WHERE student_id = $1
	`
	rows, err := r.pool.Query(ctx, q, string(studentID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.BlockProgress
	for rows.Next() {
		p, err := scanProgressRow(studentID, rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListAwaitingForBuddy returns awaiting approval rows for buddy's students.
func (r *BlockProgressRepo) ListAwaitingForBuddy(ctx context.Context, buddyID domain.UserID, page pagination.Params) ([]domain.BlockProgress, int64, error) {
	const countQ = `
		SELECT COUNT(*)
		FROM student_block_progress p
		INNER JOIN buddy_assignments a ON a.student_id = p.student_id
		WHERE a.buddy_id = $1 AND a.active = true AND a.deleted_at IS NULL
		  AND p.status = 'awaiting_approval'
	`
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, string(buddyID)).Scan(&total); err != nil {
		return nil, 0, err
	}
	const q = `
		SELECT p.student_id, p.block_id, p.status, p.submitted_at, p.approved_by, p.approved_at,
		       p.rejected_at, p.reject_reason, p.created_at, p.updated_at
		FROM student_block_progress p
		INNER JOIN buddy_assignments a ON a.student_id = p.student_id
		WHERE a.buddy_id = $1 AND a.active = true AND a.deleted_at IS NULL
		  AND p.status = 'awaiting_approval'
		ORDER BY p.submitted_at ASC NULLS LAST
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, q, string(buddyID), page.PageSize, page.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.BlockProgress
	for rows.Next() {
		var studentID string
		var blockID string
		p := domain.BlockProgress{Exists: true}
		var status string
		var approvedBy *string
		if err := rows.Scan(
			&studentID, &blockID, &status,
			&p.SubmittedAt, &approvedBy, &p.ApprovedAt,
			&p.RejectedAt, &p.RejectReason, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		p.StudentID = domain.StudentID(studentID)
		p.BlockID = domain.BlockID(blockID)
		p.Status = domain.ProgressStatus(status)
		if approvedBy != nil {
			u := domain.UserID(*approvedBy)
			p.ApprovedBy = &u
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

func scanProgressRow(studentID domain.StudentID, row pgx.Row) (domain.BlockProgress, error) {
	var blockID string
	p := domain.BlockProgress{StudentID: studentID, Exists: true}
	var status string
	var approvedBy *string
	err := row.Scan(
		&blockID, &status,
		&p.SubmittedAt, &approvedBy, &p.ApprovedAt,
		&p.RejectedAt, &p.RejectReason, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return domain.BlockProgress{}, err
	}
	p.BlockID = domain.BlockID(blockID)
	p.Status = domain.ProgressStatus(status)
	if approvedBy != nil {
		u := domain.UserID(*approvedBy)
		p.ApprovedBy = &u
	}
	return p, nil
}
