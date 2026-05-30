package domain

import (
	"encoding/json"
	"time"
)

// InAppNotification is a single inbox row for one recipient.
type InAppNotification struct {
	ID                NotificationID
	RecipientUserID   UserID
	NotificationType  NotificationType
	Title             string
	Body              string
	Payload           json.RawMessage
	ActorID           *UserID
	ReferenceType     *string
	ReferenceID       *string
	SourceOutboxID    string
	SourceEventName   string
	OccurredAt        time.Time
	ReadAt            *time.Time
	CreatedAt         time.Time
}

// IsRead reports whether the notification was marked read.
func (n InAppNotification) IsRead() bool {
	return n.ReadAt != nil
}
