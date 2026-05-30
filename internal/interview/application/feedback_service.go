package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/interview/domain"
)

// InterviewFeedbackService completes mock interviews with feedback.
type InterviewFeedbackService struct {
	repo   domain.InterviewRepository
	buddy  domain.BuddyScopePort
	tx     domain.Transactor
	events domain.EventRecorder
}

// NewInterviewFeedbackService builds InterviewFeedbackService.
func NewInterviewFeedbackService(
	repo domain.InterviewRepository,
	buddy domain.BuddyScopePort,
	tx domain.Transactor,
	events domain.EventRecorder,
) *InterviewFeedbackService {
	return &InterviewFeedbackService{repo: repo, buddy: buddy, tx: tx, events: events}
}

// CompleteMock completes mock interview with feedback.
func (s *InterviewFeedbackService) CompleteMock(ctx context.Context, buddyID, interviewID string, in FeedbackInput) (InterviewDTO, error) {
	bid, err := domain.ParseUserID(buddyID)
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
		if interview.Kind != domain.KindMock || interview.InterviewerID == nil || *interview.InterviewerID != bid {
			return domain.ErrForbidden
		}
		ok, err := s.buddy.IsActiveBuddyOf(ctx, bid, interview.StudentID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrForbidden
		}
		now := time.Now().UTC()
		if err := interview.CompleteMock(in.Feedback, in.Outcome, now); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, interview, domain.StatusScheduled); err != nil {
			return err
		}
		out = interview
		return s.events.Record(ctx, domain.EventMockCompleted, map[string]any{
			"interviewId": interviewID, "studentId": string(interview.StudentID), "buddyId": buddyID,
		})
	})
	if err != nil {
		return InterviewDTO{}, err
	}
	return toDTO(out), nil
}

// GetFeedback returns feedback for mock interview visible to student or buddy.
func (s *InterviewFeedbackService) GetFeedback(ctx context.Context, actorID, interviewID string, asBuddy bool) (string, error) {
	iid, err := domain.ParseInterviewID(interviewID)
	if err != nil {
		return "", err
	}
	interview, err := s.repo.GetByID(ctx, iid)
	if err != nil {
		return "", err
	}
	if interview.Kind != domain.KindMock {
		return "", domain.ErrNotFound
	}
	uid, err := domain.ParseUserID(actorID)
	if err != nil {
		return "", err
	}
	if asBuddy {
		if interview.InterviewerID == nil || *interview.InterviewerID != uid {
			return "", domain.ErrForbidden
		}
	} else if interview.StudentID != domain.StudentID(actorID) {
		return "", domain.ErrForbidden
	}
	if interview.Feedback == nil {
		return "", domain.ErrNotFound
	}
	return *interview.Feedback, nil
}
