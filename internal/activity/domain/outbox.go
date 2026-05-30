package domain

import (
	"encoding/json"
	"time"
)

// OutboxMessage is a row from outbox_events for ingestion.
type OutboxMessage struct {
	ID        string
	EventName string
	Payload   json.RawMessage
	CreatedAt time.Time
}
