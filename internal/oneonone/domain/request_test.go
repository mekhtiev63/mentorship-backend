package domain_test

import (
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/oneonone/domain"
)

func TestApproveFromPending(t *testing.T) {
	now := time.Now()
	admin := domain.UserID("00000000-0000-0000-0000-000000000001")
	r := domain.OneOnOneRequest{Status: domain.StatusPending}
	if err := r.Approve(admin, now, "one_on_one:x"); err != nil {
		t.Fatal(err)
	}
	if r.Status != domain.StatusAccepted {
		t.Fatalf("status %s", r.Status)
	}
}

func TestRejectRequiresReason(t *testing.T) {
	r := domain.OneOnOneRequest{Status: domain.StatusPending}
	if err := r.Reject("  ", time.Now()); err != domain.ErrRejectReason {
		t.Fatalf("got %v", err)
	}
}

func TestCancelNotPending(t *testing.T) {
	r := domain.OneOnOneRequest{Status: domain.StatusAccepted}
	if err := r.Cancel(time.Now()); err != domain.ErrInvalidTransition {
		t.Fatalf("got %v", err)
	}
}
