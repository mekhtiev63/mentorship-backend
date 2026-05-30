package domain

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// EventFilter lists query filters.
type EventFilter struct {
	From            *time.Time
	To              *time.Time
	RelatedType     *RelatedType
	IncludeCancelled bool
	IncludeDeleted  bool
}

// EventRepository persists calendar events.
type EventRepository interface {
	Insert(ctx context.Context, e CalendarEvent) error
	GetByID(ctx context.Context, id EventID) (CalendarEvent, error)
	GetForUpdate(ctx context.Context, id EventID) (CalendarEvent, error)
	Save(ctx context.Context, e CalendarEvent) error
	ListForUser(ctx context.Context, userID UserID, filter EventFilter, page pagination.Params) ([]CalendarEvent, int64, error)
	ListUpcoming(ctx context.Context, userID UserID, from time.Time, page pagination.Params) ([]CalendarEvent, int64, error)
}

// Transactor runs DB transactions.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
