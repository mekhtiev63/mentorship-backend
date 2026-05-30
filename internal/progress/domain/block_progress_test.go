package domain_test

import (
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/progress/domain"
)

func TestOnMaterialViewedFromNotStarted(t *testing.T) {
	now := time.Now().UTC()
	p := &domain.BlockProgress{Status: domain.StatusNotStarted}
	p.OnMaterialViewed(now)
	if p.Status != domain.StatusInProgress {
		t.Fatalf("status %s", p.Status)
	}
}

func TestOnMaterialViewedFromRejectedClearsDecision(t *testing.T) {
	now := time.Now().UTC()
	reason := "bad"
	p := &domain.BlockProgress{
		Status:       domain.StatusRejected,
		RejectReason: &reason,
	}
	p.OnMaterialViewed(now)
	if p.Status != domain.StatusInProgress {
		t.Fatalf("status %s", p.Status)
	}
	if p.RejectReason != nil {
		t.Fatal("expected cleared reject reason")
	}
}

func TestSubmitFromInProgress(t *testing.T) {
	now := time.Now().UTC()
	p := &domain.BlockProgress{Status: domain.StatusInProgress}
	if err := p.Submit(now); err != nil {
		t.Fatal(err)
	}
	if p.Status != domain.StatusAwaitingApproval || p.SubmittedAt == nil {
		t.Fatal("expected awaiting")
	}
}

func TestSubmitFromApprovedFails(t *testing.T) {
	p := &domain.BlockProgress{Status: domain.StatusApproved}
	if err := p.Submit(time.Now()); err != domain.ErrInvalidTransition {
		t.Fatalf("got %v", err)
	}
}

func TestApproveFromAwaiting(t *testing.T) {
	now := time.Now().UTC()
	p := &domain.BlockProgress{Status: domain.StatusAwaitingApproval}
	if err := p.Approve(domain.UserID("b"), now); err != nil {
		t.Fatal(err)
	}
	if p.Status != domain.StatusApproved {
		t.Fatal("not approved")
	}
}

func TestRejectRequiresAwaiting(t *testing.T) {
	p := &domain.BlockProgress{Status: domain.StatusInProgress}
	err := p.Reject(domain.UserID("b"), domain.RejectReason("reason"), time.Now())
	if err != domain.ErrInvalidTransition {
		t.Fatalf("got %v", err)
	}
}

func TestSequentialPolicy(t *testing.T) {
	b1 := domain.BlockID("00000000-0000-0000-0000-000000000001")
	b2 := domain.BlockID("00000000-0000-0000-0000-000000000002")
	ordered := []domain.RoadmapBlockRef{
		{BlockID: b1, SortOrder: 10},
		{BlockID: b2, SortOrder: 20},
	}
	policy := domain.SequentialBlockPolicy{}
	err := policy.CanSubmit(b2, ordered, map[domain.BlockID]domain.BlockProgress{})
	if err != domain.ErrSequentialBlock {
		t.Fatalf("expected sequential error, got %v", err)
	}
	err = policy.CanSubmit(b2, ordered, map[domain.BlockID]domain.BlockProgress{
		b1: {BlockID: b1, Status: domain.StatusApproved},
	})
	if err != nil {
		t.Fatalf("unexpected %v", err)
	}
}
