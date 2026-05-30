package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/interview/domain"
	"github.com/google/uuid"
)

// RealInterviewService handles student real interviews.
type RealInterviewService struct {
	repo   domain.InterviewRepository
	tx     domain.Transactor
	events domain.EventRecorder
}

// NewRealInterviewService builds RealInterviewService.
func NewRealInterviewService(repo domain.InterviewRepository, tx domain.Transactor, events domain.EventRecorder) *RealInterviewService {
	return &RealInterviewService{repo: repo, tx: tx, events: events}
}

// Create creates a real interview for the student.
func (s *RealInterviewService) Create(ctx context.Context, studentID string, in RealCreateInput) (InterviewDTO, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return InterviewDTO{}, err
	}
	now := time.Now().UTC()
	id := domain.InterviewID(uuid.NewString())
	interview, err := domain.NewRealInterview(id, sid, in.Company, in.Position, in.ScheduledAt.UTC(), in.StudentNotes, in.ExternalInterviewer, now)
	if err != nil {
		return InterviewDTO{}, err
	}
	if err := s.repo.Insert(ctx, interview); err != nil {
		return InterviewDTO{}, err
	}
	_ = s.events.Record(ctx, domain.EventRealCreated, map[string]any{
		"interviewId": string(id), "studentId": studentID, "company": interview.Company,
	})
	return toDTO(interview), nil
}

// Update updates own real interview in submitted state.
func (s *RealInterviewService) Update(ctx context.Context, studentID, interviewID string, in RealUpdateInput) (InterviewDTO, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return InterviewDTO{}, err
	}
	iid, err := domain.ParseInterviewID(interviewID)
	if err != nil {
		return InterviewDTO{}, err
	}
	var out domain.Interview
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		interview, err := s.repo.GetForUpdate(ctx, iid)
		if err != nil {
			return err
		}
		if interview.Kind != domain.KindReal || interview.StudentID != sid {
			return domain.ErrForbidden
		}
		now := time.Now().UTC()
		if err := interview.UpdateReal(in.Company, in.Position, in.ScheduledAt.UTC(), in.StudentNotes, in.ExternalInterviewer, now); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, interview, domain.StatusSubmitted); err != nil {
			return err
		}
		out = interview
		return s.events.Record(ctx, domain.EventRealUpdated, map[string]any{
			"interviewId": interviewID, "studentId": studentID,
		})
	})
	if err != nil {
		return InterviewDTO{}, err
	}
	return toDTO(out), nil
}

// Complete completes own real interview and publishes to catalog.
func (s *RealInterviewService) Complete(ctx context.Context, studentID, interviewID string, outcome domain.InterviewOutcome) (InterviewDTO, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return InterviewDTO{}, err
	}
	iid, err := domain.ParseInterviewID(interviewID)
	if err != nil {
		return InterviewDTO{}, err
	}
	var out domain.Interview
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		interview, err := s.repo.GetForUpdate(ctx, iid)
		if err != nil {
			return err
		}
		if interview.Kind != domain.KindReal || interview.StudentID != sid {
			return domain.ErrForbidden
		}
		now := time.Now().UTC()
		if err := interview.CompleteReal(outcome, now); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, interview, domain.StatusSubmitted); err != nil {
			return err
		}
		out = interview
		return s.events.Record(ctx, domain.EventRealCompleted, map[string]any{
			"interviewId": interviewID, "studentId": studentID, "outcome": string(outcome),
		})
	})
	if err != nil {
		return InterviewDTO{}, err
	}
	return toDTO(out), nil
}
