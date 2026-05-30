package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/interview/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// InterviewRepo implements InterviewRepository.
type InterviewRepo struct {
	pool *postgres.Pool
}

// NewInterviewRepo creates InterviewRepo.
func NewInterviewRepo(pool *postgres.Pool) *InterviewRepo {
	return &InterviewRepo{pool: pool}
}

// Insert creates interview.
func (r *InterviewRepo) Insert(ctx context.Context, i domain.Interview) error {
	id := string(i.ID)
	if id == "" {
		id = uuid.NewString()
	}
	var interviewer *string
	if i.InterviewerID != nil {
		s := string(*i.InterviewerID)
		interviewer = &s
	}
	const q = `
		INSERT INTO interviews (
			id, student_id, interviewer_id, kind, status, scheduled_at, feedback, outcome,
			company, position, student_notes, external_interviewer,
			reviewed_by, reviewed_at, cancel_reason, catalog_published,
			completed_at, cancelled_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15, $16,
			$17, $18, $19, $20
		)
	`
	_, err := db(ctx, r.pool).Exec(ctx, q,
		id, string(i.StudentID), interviewer, string(i.Kind), string(i.Status), i.ScheduledAt, i.Feedback, string(i.Outcome),
		i.Company, i.Position, i.StudentNotes, i.ExternalInterviewer,
		nil, nil, i.CancelReason, i.CatalogPublished,
		i.CompletedAt, i.CancelledAt, i.CreatedAt, i.UpdatedAt,
	)
	return err
}

// GetByID loads interview.
func (r *InterviewRepo) GetByID(ctx context.Context, id domain.InterviewID) (domain.Interview, error) {
	const q = selectInterview + ` WHERE id = $1`
	return scanInterview(db(ctx, r.pool).QueryRow(ctx, q, string(id)))
}

// GetForUpdate loads with lock.
func (r *InterviewRepo) GetForUpdate(ctx context.Context, id domain.InterviewID) (domain.Interview, error) {
	const q = selectInterview + ` WHERE id = $1 FOR UPDATE`
	return scanInterview(db(ctx, r.pool).QueryRow(ctx, q, string(id)))
}

// Save updates with optimistic status check.
func (r *InterviewRepo) Save(ctx context.Context, i domain.Interview, expected domain.InterviewStatus) error {
	var reviewedBy *string
	if i.ReviewedBy != nil {
		s := string(*i.ReviewedBy)
		reviewedBy = &s
	}
	const q = `
		UPDATE interviews SET
			status = $2, outcome = $3, scheduled_at = $4, feedback = $5,
			company = $6, position = $7, student_notes = $8, external_interviewer = $9,
			reviewed_by = $10, reviewed_at = $11, cancel_reason = $12, catalog_published = $13,
			completed_at = $14, cancelled_at = $15, updated_at = now()
		WHERE id = $1 AND status = $16
	`
	ct, err := db(ctx, r.pool).Exec(ctx, q,
		string(i.ID), string(i.Status), string(i.Outcome), i.ScheduledAt, i.Feedback,
		i.Company, i.Position, i.StudentNotes, i.ExternalInterviewer,
		reviewedBy, i.ReviewedAt, i.CancelReason, i.CatalogPublished,
		i.CompletedAt, i.CancelledAt, string(expected),
	)
	if err != nil {
		return fmt.Errorf("save interview: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrInvalidTransition
	}
	return nil
}

// ListByStudent lists student interviews.
func (r *InterviewRepo) ListByStudent(ctx context.Context, studentID domain.StudentID, kind domain.InterviewKind, status *domain.InterviewStatus, page pagination.Params) ([]domain.Interview, int64, error) {
	where := `WHERE student_id = $1 AND kind = $2`
	args := []any{string(studentID), string(kind)}
	n := 3
	if status != nil {
		where += fmt.Sprintf(` AND status = $%d`, n)
		args = append(args, string(*status))
		n++
	}
	return r.list(ctx, where, args, n, page)
}

// ListByInterviewer lists buddy mock interviews.
func (r *InterviewRepo) ListByInterviewer(ctx context.Context, interviewerID domain.UserID, kind domain.InterviewKind, status *domain.InterviewStatus, page pagination.Params) ([]domain.Interview, int64, error) {
	where := `WHERE interviewer_id = $1 AND kind = $2`
	args := []any{string(interviewerID), string(kind)}
	n := 3
	if status != nil {
		where += fmt.Sprintf(` AND status = $%d`, n)
		args = append(args, string(*status))
		n++
	}
	return r.list(ctx, where, args, n, page)
}

// ListCatalog lists published real completed interviews.
func (r *InterviewRepo) ListCatalog(ctx context.Context, page pagination.Params, company *string, outcome *domain.InterviewOutcome) ([]domain.Interview, int64, error) {
	where := `WHERE kind = 'real' AND status = 'completed' AND catalog_published = true`
	args := []any{}
	n := 1
	if company != nil && strings.TrimSpace(*company) != "" {
		where += fmt.Sprintf(` AND lower(company) LIKE lower($%d)`, n)
		args = append(args, "%"+strings.TrimSpace(*company)+"%")
		n++
	}
	if outcome != nil && *outcome != "" {
		where += fmt.Sprintf(` AND outcome = $%d`, n)
		args = append(args, string(*outcome))
		n++
	}
	return r.list(ctx, where, args, n, page)
}

func (r *InterviewRepo) list(ctx context.Context, where string, args []any, n int, page pagination.Params) ([]domain.Interview, int64, error) {
	countQ := `SELECT COUNT(*) FROM interviews ` + where
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limitPh := fmt.Sprintf(`$%d`, n)
	offsetPh := fmt.Sprintf(`$%d`, n+1)
	args = append(args, page.PageSize, page.Offset)
	q := selectInterview + ` ` + where + ` ORDER BY COALESCE(completed_at, scheduled_at, created_at) DESC LIMIT ` + limitPh + ` OFFSET ` + offsetPh
	rows, err := db(ctx, r.pool).Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.Interview
	for rows.Next() {
		item, err := scanInterview(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}
	return out, total, rows.Err()
}

const selectInterview = `
	SELECT id, student_id, interviewer_id, kind, status, scheduled_at, feedback, outcome,
	       company, position, student_notes, external_interviewer,
	       reviewed_by, reviewed_at, cancel_reason, catalog_published,
	       completed_at, cancelled_at, created_at, updated_at
	FROM interviews
`

func scanInterview(row pgx.Row) (domain.Interview, error) {
	var i domain.Interview
	var id, sid, kind, status, outcome string
	var interviewer, reviewedBy *string
	err := row.Scan(
		&id, &sid, &interviewer, &kind, &status, &i.ScheduledAt, &i.Feedback, &outcome,
		&i.Company, &i.Position, &i.StudentNotes, &i.ExternalInterviewer,
		&reviewedBy, &i.ReviewedAt, &i.CancelReason, &i.CatalogPublished,
		&i.CompletedAt, &i.CancelledAt, &i.CreatedAt, &i.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Interview{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Interview{}, err
	}
	i.ID = domain.InterviewID(id)
	i.StudentID = domain.StudentID(sid)
	i.Kind = domain.InterviewKind(kind)
	i.Status = domain.InterviewStatus(status)
	i.Outcome = domain.InterviewOutcome(outcome)
	if interviewer != nil {
		u := domain.UserID(*interviewer)
		i.InterviewerID = &u
	}
	if reviewedBy != nil {
		u := domain.UserID(*reviewedBy)
		i.ReviewedBy = &u
	}
	return i, nil
}
