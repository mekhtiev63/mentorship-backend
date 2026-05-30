package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// OutboxMessage is a pending outbox row.
type OutboxMessage struct {
	ID        string
	EventName string
	Payload   json.RawMessage
	CreatedAt time.Time
}

// BonusAccountRepository persists account balance.
type BonusAccountRepository interface {
	Get(ctx context.Context, userID UserID) (BonusAccount, error)
	GetForUpdate(ctx context.Context, userID UserID) (BonusAccount, error)
	EnsureAccount(ctx context.Context, userID UserID) error
	UpdateBalance(ctx context.Context, userID UserID, balance BonusAmount) error
}

// BonusTransactionRepository appends ledger lines.
type BonusTransactionRepository interface {
	Append(ctx context.Context, tx BonusTransaction) error
	FindByIdempotencyKey(ctx context.Context, key string) (BonusTransaction, bool, error)
	FindByReference(ctx context.Context, userID UserID, reference string) (BonusTransaction, bool, error)
	ListByUser(ctx context.Context, userID UserID, page pagination.Params) ([]BonusTransaction, int64, error)
	SumConvertDiscountPercent(ctx context.Context, userID UserID) (int, error)
}

// OutboxReader reads achievement grants from outbox.
type OutboxReader interface {
	ListPendingAchievementGranted(ctx context.Context, limit int) ([]OutboxMessage, error)
	MarkDone(ctx context.Context, id string) error
}

// EventRecorder writes outbox events.
type EventRecorder interface {
	Record(ctx context.Context, name string, payload map[string]any) error
}

// Transactor runs DB transactions.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
