package domain

import (
	"context"
	"encoding/json"

	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// RequestRepository persists one-on-one requests.
type RequestRepository interface {
	Insert(ctx context.Context, req OneOnOneRequest) error
	GetByID(ctx context.Context, id RequestID) (OneOnOneRequest, error)
	GetForUpdate(ctx context.Context, id RequestID) (OneOnOneRequest, error)
	Save(ctx context.Context, req OneOnOneRequest, expectedStatus RequestStatus) error
	ListByStudent(ctx context.Context, studentID StudentID, page pagination.Params) ([]OneOnOneRequest, int64, error)
	ListAll(ctx context.Context, page pagination.Params, status *RequestStatus) ([]OneOnOneRequest, int64, error)
}

// BuddyAssignmentPort resolves active buddy.
type BuddyAssignmentPort interface {
	GetActiveBuddyID(ctx context.Context, studentID StudentID) (BuddyID, error)
}

// BonusOneOnOnePort debits bonus for 1:1.
type BonusOneOnOnePort interface {
	HasSufficientBalance(ctx context.Context, studentID StudentID, amount int64) (bool, error)
	DebitForRequest(ctx context.Context, studentID StudentID, requestID RequestID, amount int64) error
}

// EventRecorder writes outbox events.
type EventRecorder interface {
	Record(ctx context.Context, name string, payload map[string]any) error
}

// Transactor runs DB transactions.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// PreferredSlotsJSON validates and encodes slots.
func PreferredSlotsJSON(slots json.RawMessage) ([]byte, error) {
	if len(slots) == 0 {
		return []byte("[]"), nil
	}
	if !json.Valid(slots) {
		return nil, ErrInvalidMessage
	}
	return slots, nil
}
