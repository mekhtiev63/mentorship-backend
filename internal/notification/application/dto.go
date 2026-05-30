package application

import (
	"encoding/json"
	"time"

	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// ListQuery lists inbox notifications.
type ListQuery struct {
	Page             int
	PageSize         int
	ReadStatus       *string
	NotificationType *string
}

// NotificationDTO is API-facing notification row.
type NotificationDTO struct {
	ID               string          `json:"id"`
	NotificationType string          `json:"notification_type"`
	Title            string          `json:"title"`
	Body             string          `json:"body"`
	Payload          json.RawMessage `json:"payload"`
	ActorID          *string         `json:"actor_id,omitempty"`
	ReferenceType    *string         `json:"reference_type,omitempty"`
	ReferenceID      *string         `json:"reference_id,omitempty"`
	OccurredAt       time.Time       `json:"occurred_at"`
	ReadAt           *time.Time      `json:"read_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	IsRead           bool            `json:"is_read"`
}

// ListResult is paginated inbox.
type ListResult struct {
	Items       []NotificationDTO `json:"items"`
	Meta        pagination.Meta   `json:"meta"`
	UnreadCount int64             `json:"unread_count"`
}

// UnreadCountDTO holds unread total.
type UnreadCountDTO struct {
	Count int64 `json:"count"`
}

// MarkAllReadResult reports bulk update count.
type MarkAllReadResult struct {
	Updated int64 `json:"updated"`
}
