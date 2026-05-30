package application

import (
	"context"
	"errors"
	"time"

	"github.com/go-mentorship-platform/backend/internal/finalcheck/domain"
	"github.com/google/uuid"
)

// FinalCheckQueryService reads final check status.
type FinalCheckQueryService struct {
	repo    domain.FinalAssessmentRepository
	roadmap domain.RoadmapCompletionPort
	buddy   domain.BuddyScopePort
	tx      domain.Transactor
	events  domain.EventRecorder
}

// NewFinalCheckQueryService builds FinalCheckQueryService.
func NewFinalCheckQueryService(
	repo domain.FinalAssessmentRepository,
	roadmap domain.RoadmapCompletionPort,
	buddy domain.BuddyScopePort,
	tx domain.Transactor,
	events domain.EventRecorder,
) *FinalCheckQueryService {
	return &FinalCheckQueryService{repo: repo, roadmap: roadmap, buddy: buddy, tx: tx, events: events}
}

// GetMyStatus returns student's final check status.
func (s *FinalCheckQueryService) GetMyStatus(ctx context.Context, studentID string) (FinalCheckDTO, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return FinalCheckDTO{}, err
	}
	return s.getStatus(ctx, sid, studentID, false)
}

// GetStudentStatus returns status for buddy/admin.
func (s *FinalCheckQueryService) GetStudentStatus(ctx context.Context, actorID, studentID string, isAdmin bool) (FinalCheckDTO, error) {
	if !isAdmin {
		actor, err := domain.ParseUserID(actorID)
		if err != nil {
			return FinalCheckDTO{}, domain.ErrForbidden
		}
		sid, err := domain.ParseStudentID(studentID)
		if err != nil {
			return FinalCheckDTO{}, domain.ErrForbidden
		}
		ok, err := s.buddy.IsActiveBuddyOf(ctx, actor, sid)
		if err != nil {
			return FinalCheckDTO{}, err
		}
		if !ok {
			return FinalCheckDTO{}, domain.ErrForbidden
		}
	}
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return FinalCheckDTO{}, err
	}
	return s.getStatus(ctx, sid, studentID, false)
}

func (s *FinalCheckQueryService) getStatus(ctx context.Context, sid domain.StudentID, studentID string, _ bool) (FinalCheckDTO, error) {
	programOK, err := s.roadmap.IsProgramCompleted(ctx, sid)
	if err != nil {
		return FinalCheckDTO{}, err
	}
	if !programOK {
		a, err := s.repo.GetByStudentID(ctx, sid)
		if errors.Is(err, domain.ErrNotFound) {
			return syntheticNotAvailable(sid, false), nil
		}
		if err != nil {
			return FinalCheckDTO{}, err
		}
		return toDTO(a, false), nil
	}

	var out domain.FinalAssessment
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		a, err := s.repo.GetForUpdateByStudentID(ctx, sid)
		if errors.Is(err, domain.ErrNotFound) {
			now := time.Now().UTC()
			a = domain.NewEligibleAssessment(domain.AssessmentID(uuid.NewString()), sid, now)
			if err := s.repo.Insert(ctx, a); err != nil {
				return err
			}
			_ = s.events.Record(ctx, domain.EventEligibilityGranted, map[string]any{
				"studentId": studentID, "assessmentId": string(a.ID),
			})
			out = a
			return nil
		}
		if err != nil {
			return err
		}
		expTech, expRoast := a.Tech.Status, a.Roast.Status
		if a.Tech.Status == domain.StatusNotAvailable {
			a.OpenTechAvailability(time.Now().UTC())
			if err := s.repo.Save(ctx, a, expTech, expRoast); err != nil {
				return err
			}
		}
		out = a
		return nil
	})
	if err != nil {
		return FinalCheckDTO{}, err
	}
	return toDTO(out, true), nil
}
