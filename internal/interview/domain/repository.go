package domain

import (
	"context"

	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// InterviewRepository persists interviews.
type InterviewRepository interface {
	Insert(ctx context.Context, i Interview) error
	GetByID(ctx context.Context, id InterviewID) (Interview, error)
	GetForUpdate(ctx context.Context, id InterviewID) (Interview, error)
	Save(ctx context.Context, i Interview, expected InterviewStatus) error
	ListByStudent(ctx context.Context, studentID StudentID, kind InterviewKind, status *InterviewStatus, page pagination.Params) ([]Interview, int64, error)
	ListByInterviewer(ctx context.Context, interviewerID UserID, kind InterviewKind, status *InterviewStatus, page pagination.Params) ([]Interview, int64, error)
	ListCatalog(ctx context.Context, page pagination.Params, company *string, outcome *InterviewOutcome) ([]Interview, int64, error)
}

// BuddyScopePort checks buddy assignment.
type BuddyScopePort interface {
	IsActiveBuddyOf(ctx context.Context, buddyID UserID, studentID StudentID) (bool, error)
}

// Transactor runs DB transactions.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// EventRecorder writes outbox events.
type EventRecorder interface {
	Record(ctx context.Context, name string, payload map[string]any) error
}
