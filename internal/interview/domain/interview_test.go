package domain

import (
	"testing"
	"time"
)

func TestCompleteRealPublishesCatalog(t *testing.T) {
	now := time.Now().UTC()
	i, err := NewRealInterview(InterviewID("11111111-1111-1111-1111-111111111111"), StudentID("22222222-2222-2222-2222-222222222222"),
		"Acme", "Go Dev", now, "", nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := i.CompleteReal(OutcomeOffer, now); err != nil {
		t.Fatal(err)
	}
	if !i.CatalogPublished || i.Status != StatusCompleted {
		t.Fatalf("expected completed catalog entry, got %s published=%v", i.Status, i.CatalogPublished)
	}
}

func TestCompleteMockRequiresFeedback(t *testing.T) {
	now := time.Now().UTC()
	buddy := UserID("33333333-3333-3333-3333-333333333333")
	i, _ := NewMockInterview(InterviewID("11111111-1111-1111-1111-111111111111"), StudentID("22222222-2222-2222-2222-222222222222"), buddy, now, "", now)
	if err := i.CompleteMock("", OutcomeNoResult, now); err != ErrFeedbackRequired {
		t.Fatalf("expected feedback required, got %v", err)
	}
}
