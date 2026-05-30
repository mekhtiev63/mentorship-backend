package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/oneonone/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RequestRepo implements RequestRepository.
type RequestRepo struct {
	pool *postgres.Pool
}

// NewRequestRepo creates RequestRepo.
func NewRequestRepo(pool *postgres.Pool) *RequestRepo {
	return &RequestRepo{pool: pool}
}

// Insert creates a request.
func (r *RequestRepo) Insert(ctx context.Context, req domain.OneOnOneRequest) error {
	id := string(req.ID)
	if id == "" {
		id = uuid.NewString()
	}
	const q = `
		INSERT INTO one_on_one_requests (id, student_id, buddy_id, status, message, preferred_slots)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
	`
	_, err := db(ctx, r.pool).Exec(ctx, q,
		id, string(req.StudentID), string(req.BuddyID), string(req.Status), req.Message, req.PreferredSlots,
	)
	return err
}

// GetByID loads request.
func (r *RequestRepo) GetByID(ctx context.Context, id domain.RequestID) (domain.OneOnOneRequest, error) {
	const q = `
		SELECT id, student_id, buddy_id, status, message, preferred_slots, calendar_event_id,
		       reject_reason, approved_by, approved_at, bonus_debited_at, bonus_reference,
		       cancelled_at, created_at, updated_at
		FROM one_on_one_requests WHERE id = $1
	`
	return scanRequest(db(ctx, r.pool).QueryRow(ctx, q, string(id)))
}

// GetForUpdate loads request with lock.
func (r *RequestRepo) GetForUpdate(ctx context.Context, id domain.RequestID) (domain.OneOnOneRequest, error) {
	const q = `
		SELECT id, student_id, buddy_id, status, message, preferred_slots, calendar_event_id,
		       reject_reason, approved_by, approved_at, bonus_debited_at, bonus_reference,
		       cancelled_at, created_at, updated_at
		FROM one_on_one_requests WHERE id = $1 FOR UPDATE
	`
	return scanRequest(db(ctx, r.pool).QueryRow(ctx, q, string(id)))
}

// Save updates request with optimistic status check.
func (r *RequestRepo) Save(ctx context.Context, req domain.OneOnOneRequest, expected domain.RequestStatus) error {
	const q = `
		UPDATE one_on_one_requests
		SET status = $2, reject_reason = $3, approved_by = $4, approved_at = $5,
		    bonus_debited_at = $6, bonus_reference = $7, cancelled_at = $8, updated_at = now()
		WHERE id = $1 AND status = $9
	`
	var approvedBy *string
	if req.ApprovedBy != nil {
		s := string(*req.ApprovedBy)
		approvedBy = &s
	}
	ct, err := db(ctx, r.pool).Exec(ctx, q,
		string(req.ID), string(req.Status), req.RejectReason, approvedBy, req.ApprovedAt,
		req.BonusDebitedAt, req.BonusReference, req.CancelledAt, string(expected),
	)
	if err != nil {
		return fmt.Errorf("save request: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrInvalidTransition
	}
	return nil
}

// ListByStudent lists student requests.
func (r *RequestRepo) ListByStudent(ctx context.Context, studentID domain.StudentID, page pagination.Params) ([]domain.OneOnOneRequest, int64, error) {
	const countQ = `SELECT COUNT(*) FROM one_on_one_requests WHERE student_id = $1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, string(studentID)).Scan(&total); err != nil {
		return nil, 0, err
	}
	const q = `
		SELECT id, student_id, buddy_id, status, message, preferred_slots, calendar_event_id,
		       reject_reason, approved_by, approved_at, bonus_debited_at, bonus_reference,
		       cancelled_at, created_at, updated_at
		FROM one_on_one_requests WHERE student_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`
	return r.queryList(ctx, q, total, string(studentID), page.PageSize, page.Offset)
}

// ListAll lists all requests for admin.
func (r *RequestRepo) ListAll(ctx context.Context, page pagination.Params, status *domain.RequestStatus) ([]domain.OneOnOneRequest, int64, error) {
	where := `WHERE 1=1`
	args := []any{}
	n := 1
	if status != nil {
		where += fmt.Sprintf(` AND status = $%d`, n)
		args = append(args, string(*status))
		n++
	}
	countQ := `SELECT COUNT(*) FROM one_on_one_requests ` + where
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limitPh := fmt.Sprintf(`$%d`, n)
	offsetPh := fmt.Sprintf(`$%d`, n+1)
	args = append(args, page.PageSize, page.Offset)
	q := `
		SELECT id, student_id, buddy_id, status, message, preferred_slots, calendar_event_id,
		       reject_reason, approved_by, approved_at, bonus_debited_at, bonus_reference,
		       cancelled_at, created_at, updated_at
		FROM one_on_one_requests ` + where + `
		ORDER BY created_at DESC LIMIT ` + limitPh + ` OFFSET ` + offsetPh
	return r.queryList(ctx, q, total, args...)
}

func (r *RequestRepo) queryList(ctx context.Context, q string, total int64, args ...any) ([]domain.OneOnOneRequest, int64, error) {
	rows, err := db(ctx, r.pool).Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.OneOnOneRequest
	for rows.Next() {
		req, err := scanRequest(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, req)
	}
	return out, total, rows.Err()
}

func scanRequest(row pgx.Row) (domain.OneOnOneRequest, error) {
	var req domain.OneOnOneRequest
	var id, sid, bid, status string
	var slots []byte
	var calID *string
	var approvedBy *string
	err := row.Scan(
		&id, &sid, &bid, &status, &req.Message, &slots, &calID,
		&req.RejectReason, &approvedBy, &req.ApprovedAt, &req.BonusDebitedAt, &req.BonusReference,
		&req.CancelledAt, &req.CreatedAt, &req.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.OneOnOneRequest{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.OneOnOneRequest{}, err
	}
	req.ID = domain.RequestID(id)
	req.StudentID = domain.StudentID(sid)
	req.BuddyID = domain.BuddyID(bid)
	req.Status = domain.RequestStatus(status)
	req.PreferredSlots = slots
	req.CalendarEventID = calID
	if approvedBy != nil {
		u := domain.UserID(*approvedBy)
		req.ApprovedBy = &u
	}
	return req, nil
}
