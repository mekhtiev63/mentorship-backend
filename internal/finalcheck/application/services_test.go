package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/finalcheck/application"
	"github.com/go-mentorship-platform/backend/internal/finalcheck/domain"
)

type memRepo struct {
	byStudent map[domain.StudentID]domain.FinalAssessment
}

func newMemRepo() *memRepo {
	return &memRepo{byStudent: map[domain.StudentID]domain.FinalAssessment{}}
}

func (m *memRepo) GetByStudentID(_ context.Context, sid domain.StudentID) (domain.FinalAssessment, error) {
	a, ok := m.byStudent[sid]
	if !ok {
		return domain.FinalAssessment{}, domain.ErrNotFound
	}
	return a, nil
}

func (m *memRepo) GetForUpdateByStudentID(ctx context.Context, sid domain.StudentID) (domain.FinalAssessment, error) {
	return m.GetByStudentID(ctx, sid)
}

func (m *memRepo) Insert(_ context.Context, a domain.FinalAssessment) error {
	m.byStudent[a.StudentID] = a
	return nil
}

func (m *memRepo) Save(_ context.Context, a domain.FinalAssessment, expTech, expRoast domain.TrackStatus) error {
	cur, ok := m.byStudent[a.StudentID]
	if !ok || cur.Tech.Status != expTech || cur.Roast.Status != expRoast {
		return domain.ErrInvalidTransition
	}
	m.byStudent[a.StudentID] = a
	return nil
}

type memRoadmap struct{ done bool }

func (m memRoadmap) IsProgramCompleted(context.Context, domain.StudentID) (bool, error) {
	return m.done, nil
}

type memBuddy struct{ ok bool }

func (m memBuddy) IsActiveBuddyOf(context.Context, domain.UserID, domain.StudentID) (bool, error) {
	return m.ok, nil
}

type memTx struct{}

func (memTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type memEvents struct{ names []string }

func (m *memEvents) Record(_ context.Context, name string, _ map[string]any) error {
	m.names = append(m.names, name)
	return nil
}

func TestScheduleRequiresProgram(t *testing.T) {
	svc := application.NewFinalCheckService(newMemRepo(), memRoadmap{done: false}, memBuddy{ok: true}, memTx{}, &memEvents{})
	student := "22222222-2222-2222-2222-222222222222"
	buddy := "33333333-3333-3333-3333-333333333333"
	_, err := svc.Schedule(context.Background(), buddy, student, domain.CheckTech, time.Now().UTC(), false)
	if err != domain.ErrProgramIncomplete {
		t.Fatalf("expected program incomplete, got %v", err)
	}
}

func TestBothCompletedEmitsFinalistEvent(t *testing.T) {
	repo := newMemRepo()
	events := &memEvents{}
	svc := application.NewFinalCheckService(repo, memRoadmap{done: true}, memBuddy{ok: true}, memTx{}, events)
	student := domain.StudentID("22222222-2222-2222-2222-222222222222")
	buddy := "33333333-3333-3333-3333-333333333333"
	now := time.Now().UTC()
	a := domain.NewEligibleAssessment(domain.AssessmentID("11111111-1111-1111-1111-111111111111"), student, now)
	repo.byStudent[student] = a

	_, _ = svc.Schedule(context.Background(), buddy, string(student), domain.CheckTech, now, false)
	_, _ = svc.Complete(context.Background(), buddy, string(student), domain.CheckTech, "ok", false)
	_, _ = svc.Schedule(context.Background(), buddy, string(student), domain.CheckRoast, now, false)
	_, _ = svc.Complete(context.Background(), buddy, string(student), domain.CheckRoast, "roast ok", false)

	found := false
	for _, n := range events.names {
		if n == domain.EventBothCompleted {
			found = true
		}
	}
	if !found {
		t.Fatal("expected final_check.both_completed event")
	}
}
