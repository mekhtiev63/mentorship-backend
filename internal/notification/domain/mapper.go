package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MapOutboxMessage projects outbox row to zero or more inbox notifications.
func MapOutboxMessage(ctx context.Context, msg OutboxMessage, lookup OneOnOneRequestLookup) ([]InAppNotification, error) {
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
	case SourceAchievementGranted:
		recipient, ok := userFrom(p, "userId")
		if !ok {
			return nil, nil
		}
		code := stringVal(p, "achievementCode")
		return []InAppNotification{newNotification(msg, recipient, TypeAchievementGranted,
			"Achievement unlocked", fmt.Sprintf("You earned achievement %s.", code),
			recipientPayload(p), nil, ref("achievement", code), occurred)}, nil

	case SourceProgressBlockApproved:
		recipient, ok := userFrom(p, "studentId")
		if !ok {
			return nil, nil
		}
		blockID := stringVal(p, "blockId")
		actor := userPtr(p, "approvedBy")
		return []InAppNotification{newNotification(msg, recipient, TypeBlockApproved,
			"Block approved", "Your roadmap block was approved.",
			msg.Payload, actor, ref("block", blockID), occurred)}, nil

	case SourceInterviewMockScheduled:
		recipient, ok := userFrom(p, "studentId")
		if !ok {
			return nil, nil
		}
		interviewID := stringVal(p, "interviewId")
		actor := userPtr(p, "buddyId")
		return []InAppNotification{newNotification(msg, recipient, TypeMockInterviewAssigned,
			"Mock interview scheduled", "A mock interview was scheduled for you.",
			msg.Payload, actor, ref("interview", interviewID), occurred)}, nil

	case SourceFinalCheckEligibility:
		recipient, ok := userFrom(p, "studentId")
		if !ok {
			return nil, nil
		}
		assessmentID := stringVal(p, "assessmentId")
		return []InAppNotification{newNotification(msg, recipient, TypeFinalCheckAssigned,
			"Final check available", "You can now start your final check.",
			msg.Payload, nil, ref("final_check", assessmentID), occurred)}, nil

	case SourceOneOnOneApproved:
		recipient, ok := userFrom(p, "studentId")
		if !ok {
			return nil, nil
		}
		requestID := stringVal(p, "requestId")
		actor := userPtr(p, "adminId")
		approved := newNotification(msg, recipient, TypeOneOnOneApproved,
			"1:1 request approved", "Your one-on-one request was approved.",
			msg.Payload, actor, ref("one_on_one", requestID), occurred)
		out := []InAppNotification{approved}
		if amount := int64Val(p, "amount"); amount > 0 {
			debitPayload, _ := json.Marshal(map[string]any{
				"requestId": requestID, "amount": amount, "studentId": string(recipient),
			})
			out = append(out, newNotificationWithPayload(msg, recipient, TypeBonusDebited,
				"Bonus points debited", fmt.Sprintf("%d bonus points were debited for your 1:1 request.", amount),
				debitPayload, actor, ref("bonus", requestID), occurred))
		}
		return out, nil

	case SourceOneOnOneRejected:
		requestID := stringVal(p, "requestId")
		if requestID == "" || lookup == nil {
			return nil, nil
		}
		recipient, err := lookup.StudentIDByRequestID(ctx, requestID)
		if err != nil {
			return nil, err
		}
		actor := userPtr(p, "adminId")
		return []InAppNotification{newNotification(msg, recipient, TypeOneOnOneRejected,
			"1:1 request rejected", "Your one-on-one request was rejected.",
			msg.Payload, actor, ref("one_on_one", requestID), occurred)}, nil

	case SourceOneOnOneCompleted:
		recipient, ok := userFrom(p, "studentId")
		if !ok {
			return nil, nil
		}
		requestID := stringVal(p, "requestId")
		return []InAppNotification{newNotification(msg, recipient, TypeOneOnOneCompleted,
			"1:1 session completed", "Your one-on-one session was marked completed.",
			msg.Payload, nil, ref("one_on_one", requestID), occurred)}, nil

	case SourceBonusCredited:
		recipient, ok := userFrom(p, "userId")
		if !ok {
			return nil, nil
		}
		return []InAppNotification{newNotification(msg, recipient, TypeBonusCredited,
			"Bonus credited", "Bonus points were added to your account.",
			msg.Payload, nil, ref("bonus", stringVal(p, "transactionId")), occurred)}, nil

	case SourceBonusConverted:
		recipient, ok := userFrom(p, "userId")
		if !ok {
			return nil, nil
		}
		points := int64Val(p, "points")
		body := "Bonus points were converted."
		if points > 0 {
			body = fmt.Sprintf("%d bonus points were converted.", points)
		}
		return []InAppNotification{newNotification(msg, recipient, TypeBonusDebited,
			"Bonus points debited", body,
			msg.Payload, nil, ref("bonus", stringVal(p, "transactionId")), occurred)}, nil

	default:
		return nil, nil
	}
}

func newNotification(
	msg OutboxMessage,
	recipient UserID,
	typ NotificationType,
	title, body string,
	payload json.RawMessage,
	actor *UserID,
	reference *reference,
	occurred time.Time,
) InAppNotification {
	return newNotificationWithPayload(msg, recipient, typ, title, body, payload, actor, reference, occurred)
}

func newNotificationWithPayload(
	msg OutboxMessage,
	recipient UserID,
	typ NotificationType,
	title, body string,
	payload json.RawMessage,
	actor *UserID,
	reference *reference,
	occurred time.Time,
) InAppNotification {
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	now := time.Now().UTC()
	var refType, refID *string
	if reference != nil && reference.id != "" {
		refType = &reference.typ
		refID = &reference.id
	}
	return InAppNotification{
		ID:               NotificationID(uuid.NewString()),
		RecipientUserID:  recipient,
		NotificationType: typ,
		Title:            title,
		Body:             body,
		Payload:          payload,
		ActorID:          actor,
		ReferenceType:    refType,
		ReferenceID:      refID,
		SourceOutboxID:   msg.ID,
		SourceEventName:  msg.EventName,
		OccurredAt:       occurred,
		CreatedAt:        now,
	}
}

type reference struct {
	typ, id string
}

func ref(typ, id string) *reference {
	if id == "" {
		return nil
	}
	return &reference{typ: typ, id: id}
}

func recipientPayload(p map[string]any) json.RawMessage {
	raw, _ := json.Marshal(p)
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	return raw
}

func userFrom(p map[string]any, key string) (UserID, bool) {
	s := strings.TrimSpace(stringVal(p, key))
	if s == "" {
		return "", false
	}
	u, err := ParseUserID(s)
	return u, err == nil
}

func userPtr(p map[string]any, key string) *UserID {
	u, ok := userFrom(p, key)
	if !ok {
		return nil
	}
	return &u
}

func stringVal(p map[string]any, key string) string {
	v, ok := p[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return ""
	}
}

func int64Val(p map[string]any, key string) int64 {
	v, ok := p[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case int:
		return int64(t)
	case json.Number:
		i, _ := t.Int64()
		return i
	default:
		return 0
	}
}

func parseTime(v any) (time.Time, bool) {
	s, ok := v.(string)
	if !ok || s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	return t.UTC(), err == nil
}
