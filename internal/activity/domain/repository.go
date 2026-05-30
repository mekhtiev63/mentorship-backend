package domain

import (
	"context"

	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// ActivityJournalRepository append-only activity store.
type ActivityJournalRepository interface {
	Append(ctx context.Context, e ActivityEntry) error
	GetByID(ctx context.Context, id ActivityID) (ActivityEntry, error)
	ListBySubject(ctx context.Context, subject UserID, filter ActivityFilter, page pagination.Params) ([]ActivityEntry, int64, error)
	ListAll(ctx context.Context, filter ActivityFilter, page pagination.Params) ([]ActivityEntry, int64, error)
}

// OutboxConsumerPort reads outbox for activity projection.
type OutboxConsumerPort interface {
	ListUnprocessedForActivity(ctx context.Context, limit int) ([]OutboxMessage, error)
}

// BuddyScopePort checks buddy assignment.
type BuddyScopePort interface {
	IsActiveBuddyOf(ctx context.Context, buddyID, studentID UserID) (bool, error)
}

// Transactor runs DB transactions.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
