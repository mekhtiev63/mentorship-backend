package domain_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/notification/domain"
)

const (
	studentID = "22222222-2222-2222-2222-222222222222"
	outboxID  = "11111111-1111-1111-1111-111111111111"
)

type stubLookup struct {
	student domain.UserID
}

func (s stubLookup) StudentIDByRequestID(context.Context, string) (domain.UserID, error) {
	return s.student, nil
}

func msg(name string, payload map[string]any) domain.OutboxMessage {
	raw, _ := json.Marshal(payload)
	return domain.OutboxMessage{
		ID:        outboxID,
		EventName: name,
		Payload:   raw,
		CreatedAt: time.Now().UTC(),
	}
}

func TestMapAchievementGranted(t *testing.T) {
	got, err := domain.MapOutboxMessage(context.Background(), msg(domain.SourceAchievementGranted, map[string]any{
		"userId": studentID, "achievementCode": "first_block",
	}), nil)
	if err != nil || len(got) != 1 || got[0].NotificationType != domain.TypeAchievementGranted {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if string(got[0].RecipientUserID) != studentID {
		t.Fatalf("recipient=%s", got[0].RecipientUserID)
	}
}

func TestMapOneOnOneApprovedFanOut(t *testing.T) {
	got, err := domain.MapOutboxMessage(context.Background(), msg(domain.SourceOneOnOneApproved, map[string]any{
		"studentId": studentID, "requestId": "44444444-4444-4444-4444-444444444444",
		"adminId": "55555555-5555-5555-5555-555555555555", "amount": float64(100),
	}), nil)
	if err != nil || len(got) != 2 {
		t.Fatalf("len=%d err=%v", len(got), err)
	}
	if got[0].NotificationType != domain.TypeOneOnOneApproved || got[1].NotificationType != domain.TypeBonusDebited {
		t.Fatalf("types=%s %s", got[0].NotificationType, got[1].NotificationType)
	}
}

func TestMapOneOnOneRejectedUsesLookup(t *testing.T) {
	got, err := domain.MapOutboxMessage(context.Background(), msg(domain.SourceOneOnOneRejected, map[string]any{
		"requestId": "44444444-4444-4444-4444-444444444444",
	}), stubLookup{student: domain.UserID(studentID)})
	if err != nil || len(got) != 1 || got[0].NotificationType != domain.TypeOneOnOneRejected {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestMapBonusConvertedAsDebited(t *testing.T) {
	got, err := domain.MapOutboxMessage(context.Background(), msg(domain.SourceBonusConverted, map[string]any{
		"userId": studentID, "points": float64(50),
	}), nil)
	if err != nil || len(got) != 1 || got[0].NotificationType != domain.TypeBonusDebited {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}
