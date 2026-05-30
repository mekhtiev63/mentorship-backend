package domain

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// InAppNotificationRepository persists inbox rows.
type InAppNotificationRepository interface {
	Append(ctx context.Context, n InAppNotification) error
	GetByIDForRecipient(ctx context.Context, id NotificationID, recipient UserID) (InAppNotification, error)
	ListForRecipient(ctx context.Context, recipient UserID, filter NotificationListFilter, page pagination.Params) ([]InAppNotification, int64, error)
	CountUnread(ctx context.Context, recipient UserID) (int64, error)
	MarkRead(ctx context.Context, id NotificationID, recipient UserID, readAt time.Time) error
	MarkAllRead(ctx context.Context, recipient UserID, readAt time.Time) (int64, error)
}

// OutboxConsumerPort reads outbox rows not yet receipted for notification BC.
type OutboxConsumerPort interface {
	ListUnprocessedForNotification(ctx context.Context, limit int) ([]OutboxMessage, error)
}

// OutboxReceiptRepository records processed outbox ids.
type OutboxReceiptRepository interface {
	InsertReceipt(ctx context.Context, outboxID string) error
}

// Transactor runs work in a database transaction.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(context.Context) error) error
}

// OneOnOneRequestLookup resolves student from request id (ACL).
type OneOnOneRequestLookup interface {
	StudentIDByRequestID(ctx context.Context, requestID string) (UserID, error)
}
