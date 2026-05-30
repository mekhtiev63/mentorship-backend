package domain

import (
	"testing"
	"time"
)

func TestCancelAndDelete(t *testing.T) {
	now := time.Now().UTC()
	start := now.Add(time.Hour)
	end := start.Add(time.Hour)
	e, err := NewEvent(EventID("11111111-1111-1111-1111-111111111111"), UserID("22222222-2222-2222-2222-222222222222"),
		"Meet", "", start, end, RelatedOther, nil, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Cancel(now); err != nil {
		t.Fatal(err)
	}
	if err := e.Update("x", "", start, end, nil, now); err != ErrAlreadyCancelled {
		t.Fatalf("expected cancelled error, got %v", err)
	}
}

func TestCanReadAttendee(t *testing.T) {
	now := time.Now().UTC()
	start := now.Add(time.Hour)
	end := start.Add(time.Hour)
	attendee := UserID("33333333-3333-3333-3333-333333333333")
	e, _ := NewEvent(EventID("11111111-1111-1111-1111-111111111111"), UserID("22222222-2222-2222-2222-222222222222"),
		"Meet", "", start, end, RelatedOther, nil, []UserID{attendee}, now)
	if !e.CanRead(attendee, false) {
		t.Fatal("attendee should read")
	}
}
