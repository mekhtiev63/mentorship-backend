package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/finalcheck/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// AssessmentRepo implements FinalAssessmentRepository.
type AssessmentRepo struct {
	pool *postgres.Pool
}

// NewAssessmentRepo creates AssessmentRepo.
func NewAssessmentRepo(pool *postgres.Pool) *AssessmentRepo {
	return &AssessmentRepo{pool: pool}
}

const selectAssessment = `
	SELECT id, student_id,
	       tech_status, tech_reviewer_id, tech_scheduled_at, tech_feedback, tech_completed_at, tech_failed_at, tech_fail_reason,
	       roast_status, roast_reviewer_id, roast_scheduled_at, roast_feedback, roast_completed_at, roast_failed_at, roast_fail_reason,
	       finalist_event_emitted, created_at, updated_at
	FROM final_assessments
`

// GetByStudentID loads assessment.
func (r *AssessmentRepo) GetByStudentID(ctx context.Context, studentID domain.StudentID) (domain.FinalAssessment, error) {
	const q = selectAssessment + ` WHERE student_id = $1`
	return scanAssessment(db(ctx, r.pool).QueryRow(ctx, q, string(studentID)))
}

// GetForUpdateByStudentID loads with lock.
func (r *AssessmentRepo) GetForUpdateByStudentID(ctx context.Context, studentID domain.StudentID) (domain.FinalAssessment, error) {
	const q = selectAssessment + ` WHERE student_id = $1 FOR UPDATE`
	return scanAssessment(db(ctx, r.pool).QueryRow(ctx, q, string(studentID)))
}

// Insert creates assessment.
func (r *AssessmentRepo) Insert(ctx context.Context, a domain.FinalAssessment) error {
	id := string(a.ID)
	if id == "" {
		id = uuid.NewString()
	}
	const q = `
		INSERT INTO final_assessments (
			id, student_id,
			tech_status, tech_reviewer_id, tech_scheduled_at, tech_feedback, tech_completed_at, tech_failed_at, tech_fail_reason,
			roast_status, roast_reviewer_id, roast_scheduled_at, roast_feedback, roast_completed_at, roast_failed_at, roast_fail_reason,
			finalist_event_emitted, created_at, updated_at
		) VALUES (
			$1, $2,
			$3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19
		)
	`
	tRev, rRev := reviewerStr(a.Tech.ReviewerID), reviewerStr(a.Roast.ReviewerID)
	_, err := db(ctx, r.pool).Exec(ctx, q,
		id, string(a.StudentID),
		string(a.Tech.Status), tRev, a.Tech.ScheduledAt, a.Tech.Feedback, a.Tech.CompletedAt, a.Tech.FailedAt, a.Tech.FailReason,
		string(a.Roast.Status), rRev, a.Roast.ScheduledAt, a.Roast.Feedback, a.Roast.CompletedAt, a.Roast.FailedAt, a.Roast.FailReason,
		a.FinalistEventEmitted, a.CreatedAt, a.UpdatedAt,
	)
	return err
}

// Save updates with optimistic track status checks.
func (r *AssessmentRepo) Save(ctx context.Context, a domain.FinalAssessment, expectedTech, expectedRoast domain.TrackStatus) error {
	const q = `
		UPDATE final_assessments SET
			tech_status = $2, tech_reviewer_id = $3, tech_scheduled_at = $4, tech_feedback = $5,
			tech_completed_at = $6, tech_failed_at = $7, tech_fail_reason = $8,
			roast_status = $9, roast_reviewer_id = $10, roast_scheduled_at = $11, roast_feedback = $12,
			roast_completed_at = $13, roast_failed_at = $14, roast_fail_reason = $15,
			finalist_event_emitted = $16, updated_at = now()
		WHERE student_id = $1 AND tech_status = $17 AND roast_status = $18
	`
	tRev, rRev := reviewerStr(a.Tech.ReviewerID), reviewerStr(a.Roast.ReviewerID)
	ct, err := db(ctx, r.pool).Exec(ctx, q,
		string(a.StudentID),
		string(a.Tech.Status), tRev, a.Tech.ScheduledAt, a.Tech.Feedback, a.Tech.CompletedAt, a.Tech.FailedAt, a.Tech.FailReason,
		string(a.Roast.Status), rRev, a.Roast.ScheduledAt, a.Roast.Feedback, a.Roast.CompletedAt, a.Roast.FailedAt, a.Roast.FailReason,
		a.FinalistEventEmitted,
		string(expectedTech), string(expectedRoast),
	)
	if err != nil {
		return fmt.Errorf("save assessment: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrInvalidTransition
	}
	return nil
}

func reviewerStr(id *domain.UserID) *string {
	if id == nil {
		return nil
	}
	s := string(*id)
	return &s
}

func scanAssessment(row pgx.Row) (domain.FinalAssessment, error) {
	var a domain.FinalAssessment
	var id, sid string
	var techSt, roastSt string
	var tRev, rRev *string
	err := row.Scan(
		&id, &sid,
		&techSt, &tRev, &a.Tech.ScheduledAt, &a.Tech.Feedback, &a.Tech.CompletedAt, &a.Tech.FailedAt, &a.Tech.FailReason,
		&roastSt, &rRev, &a.Roast.ScheduledAt, &a.Roast.Feedback, &a.Roast.CompletedAt, &a.Roast.FailedAt, &a.Roast.FailReason,
		&a.FinalistEventEmitted, &a.CreatedAt, &a.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.FinalAssessment{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.FinalAssessment{}, err
	}
	a.ID = domain.AssessmentID(id)
	a.StudentID = domain.StudentID(sid)
	a.Tech.Status = domain.TrackStatus(techSt)
	a.Roast.Status = domain.TrackStatus(roastSt)
	if tRev != nil {
		u := domain.UserID(*tRev)
		a.Tech.ReviewerID = &u
	}
	if rRev != nil {
		u := domain.UserID(*rRev)
		a.Roast.ReviewerID = &u
	}
	return a, nil
}
