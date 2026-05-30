package domain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MapOutboxMessage maps outbox row to activity entry when recognized.
func MapOutboxMessage(msg OutboxMessage) (ActivityEntry, bool) {
	var p map[string]any
	if len(msg.Payload) > 0 {
		_ = json.Unmarshal(msg.Payload, &p)
	}
	if p == nil {
		p = map[string]any{}
	}
	occurred := msg.CreatedAt.UTC()
	if t, ok := parseTime(p["occurredAt"]); ok {
		occurred = t
	} else if t, ok := parseTime(p["completedAt"]); ok {
		occurred = t
	}

	switch msg.EventName {
	case SourceProgressMaterialViewed:
		sub, ok := subjectFrom(p, "studentId")
		if !ok {
			return ActivityEntry{}, false
		}
		return newEntry(msg, sub, ptrUser(stringVal(p, "studentId")), TypeMaterialViewed, "viewed", "material", stringPtr(p, "materialId"), occurred), true
	case SourceProgressBlockApproved:
		sub, ok := subjectFrom(p, "studentId")
		if !ok {
			return ActivityEntry{}, false
		}
		actor := ptrUser(stringVal(p, "approvedBy"))
		return newEntry(msg, sub, actor, TypeBlockApproved, "approved", "block", stringPtr(p, "blockId"), occurred), true
	case SourceAchievementGranted:
		sub, ok := subjectFrom(p, "userId")
		if !ok {
			return ActivityEntry{}, false
		}
		return newEntry(msg, sub, nil, TypeAchievementGranted, "granted", "achievement", stringPtr(p, "achievementCode"), occurred), true
	case SourceBonusCredited:
		sub, ok := subjectFrom(p, "userId")
		if !ok {
			return ActivityEntry{}, false
		}
		p["bonusKind"] = "credit"
		return newEntryWithPayload(msg, sub, nil, TypeBonusTransaction, "credited", "bonus", nil, occurred, p), true
	case SourceBonusConverted:
		sub, ok := subjectFrom(p, "userId")
		if !ok {
			return ActivityEntry{}, false
		}
		p["bonusKind"] = "convert"
		return newEntryWithPayload(msg, sub, nil, TypeBonusTransaction, "converted", "bonus", nil, occurred, p), true
	case SourceOneOnOneApproved:
		sub, ok := subjectFrom(p, "studentId")
		if !ok {
			return ActivityEntry{}, false
		}
		p["bonusKind"] = "debit"
		actor := ptrUser(stringVal(p, "adminId"))
		return newEntryWithPayload(msg, sub, actor, TypeOneOnOneApproved, "approved", "one_on_one", stringPtr(p, "requestId"), occurred, p), true
	case SourceInterviewRealCreated, SourceInterviewMockScheduled:
		sub, ok := subjectFrom(p, "studentId")
		if !ok {
			return ActivityEntry{}, false
		}
		obj := stringPtr(p, "interviewId")
		return newEntry(msg, sub, ptrUser(stringVal(p, "studentId")), TypeInterviewCreated, "created", "interview", obj, occurred), true
	case SourceInterviewRealUpdated:
		sub, ok := subjectFrom(p, "studentId")
		if !ok {
			return ActivityEntry{}, false
		}
		return newEntry(msg, sub, ptrUser(stringVal(p, "studentId")), TypeInterviewUpdated, "updated", "interview", stringPtr(p, "interviewId"), occurred), true
	case SourceFinalCheckBothCompleted:
		sub, ok := subjectFrom(p, "studentId")
		if !ok {
			return ActivityEntry{}, false
		}
		return newEntry(msg, sub, nil, TypeFinalCheckCompleted, "completed", "final_check", stringPtr(p, "assessmentId"), occurred), true
	case SourceCalendarEventCreated:
		sub, ok := subjectFrom(p, "organizerId")
		if !ok {
			sub, ok = subjectFrom(p, "userId")
		}
		if !ok {
			return ActivityEntry{}, false
		}
		actor := ptrUser(stringVal(p, "organizerId"))
		return newEntry(msg, sub, actor, TypeCalendarEventCreated, "created", "calendar_event", stringPtr(p, "eventId"), occurred), true
	default:
		return ActivityEntry{}, false
	}
}

func newEntry(msg OutboxMessage, subject UserID, actor *UserID, typ ActivityType, verb, objectType string, objectID *string, occurred time.Time) ActivityEntry {
	return newEntryWithPayload(msg, subject, actor, typ, verb, objectType, objectID, occurred, nil)
}

func newEntryWithPayload(msg OutboxMessage, subject UserID, actor *UserID, typ ActivityType, verb, objectType string, objectID *string, occurred time.Time, extra map[string]any) ActivityEntry {
	raw := msg.Payload
	if extra != nil {
		raw, _ = json.Marshal(extra)
	}
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	return ActivityEntry{
		ID:              ActivityID(uuid.NewString()),
		SubjectUserID:   subject,
		ActorID:         actor,
		ActivityType:    typ,
		Verb:            verb,
		ObjectType:      objectType,
		ObjectID:        objectID,
		Payload:         raw,
		SourceOutboxID:  msg.ID,
		SourceEventName: msg.EventName,
		OccurredAt:      occurred,
		CreatedAt:       time.Now().UTC(),
	}
}

func subjectFrom(p map[string]any, key string) (UserID, bool) {
	s := strings.TrimSpace(stringVal(p, key))
	if s == "" {
		return "", false
	}
	u, err := ParseUserID(s)
	return u, err == nil
}

func stringVal(p map[string]any, key string) string {
	v, ok := p[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}

func stringPtr(p map[string]any, key string) *string {
	s := stringVal(p, key)
	if s == "" {
		return nil
	}
	return &s
}

func ptrUser(s string) *UserID {
	if s == "" {
		return nil
	}
	u := UserID(s)
	return &u
}

func parseTime(v any) (time.Time, bool) {
	s, ok := v.(string)
	if !ok || s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	return t.UTC(), err == nil
}
