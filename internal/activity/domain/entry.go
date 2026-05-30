package domain

import (
	"encoding/json"
	"time"
)

// ActivityEntry is an immutable journal row.
type ActivityEntry struct {
	ID              ActivityID
	SubjectUserID   UserID
	ActorID         *UserID
	ActivityType    ActivityType
	Verb            string
	ObjectType      string
	ObjectID        *string
	Payload         json.RawMessage
	SourceOutboxID  string
	SourceEventName string
	OccurredAt      time.Time
	CreatedAt       time.Time
}
