package application

import (
	"context"
	"errors"
	"time"

	"github.com/go-mentorship-platform/backend/internal/finalcheck/domain"
	"github.com/google/uuid"
)

// FinalCheckService schedules, completes, and fails final checks.
type FinalCheckService struct {
	repo    domain.FinalAssessmentRepository
	roadmap domain.RoadmapCompletionPort
	buddy   domain.BuddyScopePort
	tx      domain.Transactor
	events  domain.EventRecorder
}

// NewFinalCheckService builds FinalCheckService.
func NewFinalCheckService(
	repo domain.FinalAssessmentRepository,
	roadmap domain.RoadmapCompletionPort,
	buddy domain.BuddyScopePort,
	tx domain.Transactor,
	events domain.EventRecorder,
) *FinalCheckService {
	return &FinalCheckService{repo: repo, roadmap: roadmap, buddy: buddy, tx: tx, events: events}
}

// Schedule assigns a final check track.
func (s *FinalCheckService) Schedule(ctx context.Context, actorID, studentID string, kind domain.CheckKind, scheduledAt time.Time, isAdmin bool) (FinalCheckDTO, error) {
	if err := s.ensureCanManage(ctx, actorID, studentID, isAdmin); err != nil {
		return FinalCheckDTO{}, err
	}
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return FinalCheckDTO{}, err
	}
	reviewer, err := domain.ParseUserID(actorID)
	if err != nil {
		return FinalCheckDTO{}, err
	}
	programOK, err := s.roadmap.IsProgramCompleted(ctx, sid)
	if err != nil {
		return FinalCheckDTO{}, err
	}
	if !programOK {
		return FinalCheckDTO{}, domain.ErrProgramIncomplete
	}

	var out domain.FinalAssessment
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		a, expTech, expRoast, err := s.loadForUpdate(ctx, sid, programOK)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := a.Schedule(kind, reviewer, scheduledAt.UTC(), now); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, a, expTech, expRoast); err != nil {
			return err
		}
		out = a
		return s.recordTrackEvent(ctx, kind, "scheduled", a, reviewer, scheduledAt)
	})
	if err != nil {
		return FinalCheckDTO{}, err
	}
	return toDTO(out, true), nil
}

// Complete confirms a scheduled track.
func (s *FinalCheckService) Complete(ctx context.Context, actorID, studentID string, kind domain.CheckKind, feedback string, isAdmin bool) (FinalCheckDTO, error) {
	if err := s.ensureCanManage(ctx, actorID, studentID, isAdmin); err != nil {
		return FinalCheckDTO{}, err
	}
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return FinalCheckDTO{}, err
	}
	reviewer, err := domain.ParseUserID(actorID)
	if err != nil {
		return FinalCheckDTO{}, err
	}

	var out domain.FinalAssessment
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		programOK, err := s.roadmap.IsProgramCompleted(ctx, sid)
		if err != nil {
			return err
		}
		if !programOK {
			return domain.ErrProgramIncomplete
		}
		a, expTech, expRoast, err := s.loadForUpdate(ctx, sid, programOK)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := a.Complete(kind, feedback, now); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, a, expTech, expRoast); err != nil {
			return err
		}
		out = a
		if err := s.recordTrackEvent(ctx, kind, "completed", a, reviewer, now); err != nil {
			return err
		}
		if kind == domain.CheckTech && a.Roast.Status == domain.StatusAvailable {
			_ = s.events.Record(ctx, domain.EventRoastAvailable, map[string]any{
				"studentId": studentID, "assessmentId": string(a.ID),
			})
		}
		return s.emitFinalistIfNeeded(ctx, a, studentID, now)
	})
	if err != nil {
		return FinalCheckDTO{}, err
	}
	programOK, _ := s.roadmap.IsProgramCompleted(ctx, sid)
	return toDTO(out, programOK), nil
}

// Fail marks a scheduled track as failed (terminal).
func (s *FinalCheckService) Fail(ctx context.Context, actorID, studentID string, kind domain.CheckKind, reason string, isAdmin bool) (FinalCheckDTO, error) {
	if err := s.ensureCanManage(ctx, actorID, studentID, isAdmin); err != nil {
		return FinalCheckDTO{}, err
	}
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return FinalCheckDTO{}, err
	}
	reviewer, err := domain.ParseUserID(actorID)
	if err != nil {
		return FinalCheckDTO{}, err
	}

	var out domain.FinalAssessment
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		programOK, err := s.roadmap.IsProgramCompleted(ctx, sid)
		if err != nil {
			return err
		}
		if !programOK {
			return domain.ErrProgramIncomplete
		}
		a, expTech, expRoast, err := s.loadForUpdate(ctx, sid, programOK)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := a.Fail(kind, reason, now); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, a, expTech, expRoast); err != nil {
			return err
		}
		out = a
		return s.recordTrackEvent(ctx, kind, "failed", a, reviewer, now)
	})
	if err != nil {
		return FinalCheckDTO{}, err
	}
	programOK, _ := s.roadmap.IsProgramCompleted(ctx, sid)
	return toDTO(out, programOK), nil
}

func (s *FinalCheckService) ensureCanManage(ctx context.Context, actorID, studentID string, isAdmin bool) error {
	if isAdmin {
		return nil
	}
	actor, err := domain.ParseUserID(actorID)
	if err != nil {
		return domain.ErrForbidden
	}
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return domain.ErrForbidden
	}
	ok, err := s.buddy.IsActiveBuddyOf(ctx, actor, sid)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func (s *FinalCheckService) loadForUpdate(ctx context.Context, sid domain.StudentID, programOK bool) (domain.FinalAssessment, domain.TrackStatus, domain.TrackStatus, error) {
	a, err := s.repo.GetForUpdateByStudentID(ctx, sid)
	if errors.Is(err, domain.ErrNotFound) {
		if !programOK {
			return domain.FinalAssessment{}, "", "", domain.ErrProgramIncomplete
		}
		now := time.Now().UTC()
		a = domain.NewEligibleAssessment(domain.AssessmentID(uuid.NewString()), sid, now)
		if err := s.repo.Insert(ctx, a); err != nil {
			return domain.FinalAssessment{}, "", "", err
		}
		_ = s.events.Record(ctx, domain.EventEligibilityGranted, map[string]any{
			"studentId": string(sid), "assessmentId": string(a.ID),
		})
		return a, domain.StatusAvailable, domain.StatusNotAvailable, nil
	}
	if err != nil {
		return domain.FinalAssessment{}, "", "", err
	}
	expTech, expRoast := a.Tech.Status, a.Roast.Status
	if programOK && a.Tech.Status == domain.StatusNotAvailable {
		now := time.Now().UTC()
		a.OpenTechAvailability(now)
		if err := s.repo.Save(ctx, a, expTech, expRoast); err != nil {
			return domain.FinalAssessment{}, "", "", err
		}
		expTech = a.Tech.Status
	}
	return a, expTech, expRoast, nil
}

func (s *FinalCheckService) recordTrackEvent(ctx context.Context, kind domain.CheckKind, action string, a domain.FinalAssessment, reviewer domain.UserID, at time.Time) error {
	base := map[string]any{
		"studentId": string(a.StudentID), "assessmentId": string(a.ID), "reviewerId": string(reviewer),
	}
	switch kind {
	case domain.CheckTech:
		switch action {
		case "scheduled":
			base["scheduledAt"] = at
			return s.events.Record(ctx, domain.EventTechScheduled, base)
		case "completed":
			base["completedAt"] = at
			return s.events.Record(ctx, domain.EventTechCompleted, base)
		case "failed":
			return s.events.Record(ctx, domain.EventTechFailed, base)
		}
	case domain.CheckRoast:
		switch action {
		case "scheduled":
			base["scheduledAt"] = at
			return s.events.Record(ctx, domain.EventRoastScheduled, base)
		case "completed":
			base["completedAt"] = at
			return s.events.Record(ctx, domain.EventRoastCompleted, base)
		case "failed":
			return s.events.Record(ctx, domain.EventRoastFailed, base)
		}
	}
	return nil
}

func (s *FinalCheckService) emitFinalistIfNeeded(ctx context.Context, a domain.FinalAssessment, studentID string, now time.Time) error {
	if !a.BothCompleted() || a.FinalistEventEmitted {
		return nil
	}
	a.FinalistEventEmitted = true
	if err := s.repo.Save(ctx, a, a.Tech.Status, a.Roast.Status); err != nil {
		return err
	}
	return s.events.Record(ctx, domain.EventBothCompleted, map[string]any{
		"studentId": studentID, "assessmentId": string(a.ID),
		"achievementCode": domain.AchievementFinalistCode, "completedAt": now,
	})
}
