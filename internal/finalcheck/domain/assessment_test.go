package domain

import (
	"testing"
	"time"
)

func TestTechCompleteOpensRoast(t *testing.T) {
	now := time.Now().UTC()
	a := NewEligibleAssessment(AssessmentID("11111111-1111-1111-1111-111111111111"), StudentID("22222222-2222-2222-2222-222222222222"), now)
	reviewer := UserID("33333333-3333-3333-3333-333333333333")
	_ = a.Schedule(CheckTech, reviewer, now, now)
	if err := a.Complete(CheckTech, "good", now); err != nil {
		t.Fatal(err)
	}
	if a.Roast.Status != StatusAvailable {
		t.Fatalf("roast expected available, got %s", a.Roast.Status)
	}
}

func TestFailIsTerminal(t *testing.T) {
	now := time.Now().UTC()
	a := NewEligibleAssessment(AssessmentID("11111111-1111-1111-1111-111111111111"), StudentID("22222222-2222-2222-2222-222222222222"), now)
	reviewer := UserID("33333333-3333-3333-3333-333333333333")
	_ = a.Schedule(CheckTech, reviewer, now, now)
	_ = a.Fail(CheckTech, "not ready", now)
	if err := a.Schedule(CheckTech, reviewer, now, now); err != ErrRetryNotAllowed {
		t.Fatalf("expected retry not allowed, got %v", err)
	}
}
