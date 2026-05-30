package application_test

import (
	"context"
	"testing"

	"github.com/go-mentorship-platform/backend/internal/activity/application"
	"github.com/go-mentorship-platform/backend/internal/activity/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

type memJournal struct {
	entries []domain.ActivityEntry
}

func (m *memJournal) Append(_ context.Context, e domain.ActivityEntry) error {
	for _, x := range m.entries {
		if x.SourceOutboxID == e.SourceOutboxID {
			return domain.ErrDuplicate
		}
	}
	m.entries = append(m.entries, e)
	return nil
}

func (m *memJournal) GetByID(context.Context, domain.ActivityID) (domain.ActivityEntry, error) {
	return domain.ActivityEntry{}, domain.ErrNotFound
}

func (m *memJournal) ListBySubject(_ context.Context, subject domain.UserID, _ domain.ActivityFilter, _ pagination.Params) ([]domain.ActivityEntry, int64, error) {
	var out []domain.ActivityEntry
	for _, e := range m.entries {
		if e.SubjectUserID == subject {
			out = append(out, e)
		}
	}
	return out, int64(len(out)), nil
}

func (m *memJournal) ListAll(context.Context, domain.ActivityFilter, pagination.Params) ([]domain.ActivityEntry, int64, error) {
	return m.entries, int64(len(m.entries)), nil
}

type memOutbox struct{ msgs []domain.OutboxMessage }

func (m *memOutbox) ListUnprocessedForActivity(context.Context, int) ([]domain.OutboxMessage, error) {
	return m.msgs, nil
}

type memTx struct{}

func (memTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type memBuddy struct{ ok bool }

func (m memBuddy) IsActiveBuddyOf(context.Context, domain.UserID, domain.UserID) (bool, error) {
	return m.ok, nil
}

func TestActivityServiceIngest(t *testing.T) {
	j := &memJournal{}
	payload := []byte(`{"studentId":"22222222-2222-2222-2222-222222222222","materialId":"33333333-3333-3333-3333-333333333333"}`)
	ob := &memOutbox{msgs: []domain.OutboxMessage{{
		ID: "11111111-1111-1111-1111-111111111111",
		EventName: domain.SourceProgressMaterialViewed,
		Payload: payload,
	}}}
	svc := application.NewActivityService(j, ob, memTx{}, nil)
	n, err := svc.ProcessPending(context.Background(), 10)
	if err != nil || n != 1 || len(j.entries) != 1 {
		t.Fatalf("n=%d err=%v entries=%d", n, err, len(j.entries))
	}
}

func TestBuddyFeedForbidden(t *testing.T) {
	q := application.NewActivityQueryService(&memJournal{}, memBuddy{ok: false})
	_, err := q.ListStudentActivity(context.Background(),
		"33333333-3333-3333-3333-333333333333",
		"22222222-2222-2222-2222-222222222222",
		application.ListQuery{Page: 1, PageSize: 20},
	)
	if err != domain.ErrForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
