package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/calendar/application"
	"github.com/go-mentorship-platform/backend/internal/calendar/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

type memRepo struct {
	events map[domain.EventID]domain.CalendarEvent
}

func newMemRepo() *memRepo {
	return &memRepo{events: map[domain.EventID]domain.CalendarEvent{}}
}

func (m *memRepo) Insert(_ context.Context, e domain.CalendarEvent) error {
	m.events[e.ID] = e
	return nil
}

func (m *memRepo) GetByID(_ context.Context, id domain.EventID) (domain.CalendarEvent, error) {
	e, ok := m.events[id]
	if !ok {
		return domain.CalendarEvent{}, domain.ErrNotFound
	}
	return e, nil
}

func (m *memRepo) GetForUpdate(ctx context.Context, id domain.EventID) (domain.CalendarEvent, error) {
	return m.GetByID(ctx, id)
}

func (m *memRepo) Save(_ context.Context, e domain.CalendarEvent) error {
	m.events[e.ID] = e
	return nil
}

func (m *memRepo) ListForUser(_ context.Context, userID domain.UserID, _ domain.EventFilter, _ pagination.Params) ([]domain.CalendarEvent, int64, error) {
	var out []domain.CalendarEvent
	for _, e := range m.events {
		if e.CanRead(userID, false) {
			out = append(out, e)
		}
	}
	return out, int64(len(out)), nil
}

func (m *memRepo) ListUpcoming(ctx context.Context, userID domain.UserID, from time.Time, page pagination.Params) ([]domain.CalendarEvent, int64, error) {
	return m.ListForUser(ctx, userID, domain.EventFilter{}, page)
}

type memTx struct{}

func (memTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestUpdateForbiddenForAttendee(t *testing.T) {
	repo := newMemRepo()
	now := time.Now().UTC()
	start := now.Add(time.Hour)
	end := start.Add(time.Hour)
	org := domain.UserID("22222222-2222-2222-2222-222222222222")
	att := domain.UserID("33333333-3333-3333-3333-333333333333")
	id := domain.EventID("11111111-1111-1111-1111-111111111111")
	e, _ := domain.NewEvent(id, org, "T", "", start, end, domain.RelatedOther, nil, []domain.UserID{att}, now)
	repo.events[id] = e

	svc := application.NewCalendarEventService(repo, memTx{})
	_, err := svc.Update(context.Background(), string(att), string(id), application.UpdateEventInput{
		Title: "X", StartsAt: start, EndsAt: end,
	}, false)
	if err != domain.ErrForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestCancelEvent(t *testing.T) {
	repo := newMemRepo()
	now := time.Now().UTC()
	start := now.Add(time.Hour)
	end := start.Add(time.Hour)
	org := domain.UserID("22222222-2222-2222-2222-222222222222")
	id := domain.EventID("11111111-1111-1111-1111-111111111111")
	e, _ := domain.NewEvent(id, org, "T", "", start, end, domain.RelatedOther, nil, nil, now)
	repo.events[id] = e

	svc := application.NewCalendarEventService(repo, memTx{})
	dto, err := svc.Cancel(context.Background(), string(org), string(id), false)
	if err != nil {
		t.Fatal(err)
	}
	if !dto.Cancelled {
		t.Fatal("expected cancelled")
	}
}
