package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMapMaterialViewed(t *testing.T) {
	payload, _ := json.Marshal(map[string]any{
		"studentId": "22222222-2222-2222-2222-222222222222",
		"materialId": "33333333-3333-3333-3333-333333333333",
	})
	now := time.Now().UTC()
	entry, ok := MapOutboxMessage(OutboxMessage{
		ID: "11111111-1111-1111-1111-111111111111",
		EventName: SourceProgressMaterialViewed,
		Payload: payload,
		CreatedAt: now,
	})
	if !ok || entry.ActivityType != TypeMaterialViewed {
		t.Fatalf("unexpected %+v ok=%v", entry, ok)
	}
}

func TestMapUnknownSkipped(t *testing.T) {
	_, ok := MapOutboxMessage(OutboxMessage{EventName: "roadmap.block.published"})
	if ok {
		t.Fatal("expected skip")
	}
}
