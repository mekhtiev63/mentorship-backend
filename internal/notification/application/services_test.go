package application_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/notification/application"
	"github.com/go-mentorship-platform/backend/internal/notification/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

type memInbox struct {
	items []domain.InAppNotification
}

func (m *memInbox) Append(_ context.Context, n domain.InAppNotification) error {
	for _, x := range m.items {
		if x.SourceOutboxID == n.SourceOutboxID &&
			x.RecipientUserID == n.RecipientUserID &&
			x.NotificationType == n.NotificationType {
			return domain.ErrDuplicate
		}
	}
	m.items = append(m.items, n)
	return nil
}

func (m *memInbox) GetByIDForRecipient(_ context.Context, id domain.NotificationID, recipient domain.UserID) (domain.InAppNotification, error) {
	for _, x := range m.items {
		if x.ID == id && x.RecipientUserID == recipient {
			return x, nil
		}
	}
	return domain.InAppNotification{}, domain.ErrNotFound
}

func (m *memInbox) ListForRecipient(_ context.Context, recipient domain.UserID, filter domain.NotificationListFilter, page pagination.Params) ([]domain.InAppNotification, int64, error) {
	var out []domain.InAppNotification
	for _, x := range m.items {
		if x.RecipientUserID != recipient {
			continue
		}
		if filter.ReadStatus != nil {
			unread := x.ReadAt == nil
			if *filter.ReadStatus == domain.ReadStatusUnread && !unread {
				continue
			}
			if *filter.ReadStatus == domain.ReadStatusRead && unread {
				continue
			}
		}
		out = append(out, x)
	}
	total := int64(len(out))
	if page.Offset >= len(out) {
		return nil, total, nil
	}
	end := page.Offset + page.PageSize
	if end > len(out) {
		end = len(out)
	}
	return out[page.Offset:end], total, nil
}

func (m *memInbox) CountUnread(_ context.Context, recipient domain.UserID) (int64, error) {
	var n int64
	for _, x := range m.items {
		if x.RecipientUserID == recipient && x.ReadAt == nil {
			n++
		}
	}
	return n, nil
}

func (m *memInbox) MarkRead(_ context.Context, id domain.NotificationID, recipient domain.UserID, readAt time.Time) error {
	for i, x := range m.items {
		if x.ID == id && x.RecipientUserID == recipient && x.ReadAt == nil {
			m.items[i].ReadAt = &readAt
		}
	}
	return nil
}

func (m *memInbox) MarkAllRead(_ context.Context, recipient domain.UserID, readAt time.Time) (int64, error) {
	var n int64
	for i, x := range m.items {
		if x.RecipientUserID == recipient && x.ReadAt == nil {
			m.items[i].ReadAt = &readAt
			n++
		}
	}
	return n, nil
}

type memOutbox struct{ msgs []domain.OutboxMessage }

func (m *memOutbox) ListUnprocessedForNotification(context.Context, int) ([]domain.OutboxMessage, error) {
	return m.msgs, nil
}

type memReceipt struct{ ids []string }

func (m *memReceipt) InsertReceipt(_ context.Context, id string) error {
	m.ids = append(m.ids, id)
	return nil
}

type memLookup struct{}

func (memLookup) StudentIDByRequestID(context.Context, string) (domain.UserID, error) {
	return "22222222-2222-2222-2222-222222222222", nil
}

type memTx struct{}

func (memTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestNotificationServiceIngest(t *testing.T) {
	inbox := &memInbox{}
	receipt := &memReceipt{}
	payload, _ := json.Marshal(map[string]any{
		"userId": "22222222-2222-2222-2222-222222222222", "achievementCode": "x",
	})
	ob := &memOutbox{msgs: []domain.OutboxMessage{{
		ID: "11111111-1111-1111-1111-111111111111", EventName: domain.SourceAchievementGranted, Payload: payload,
	}}}
	svc := application.NewNotificationService(inbox, ob, receipt, memLookup{}, memTx{}, nil)
	n, err := svc.ProcessPending(context.Background(), 10)
	if err != nil || n != 1 || len(inbox.items) != 1 || len(receipt.ids) != 1 {
		t.Fatalf("n=%d err=%v items=%d receipts=%d", n, err, len(inbox.items), len(receipt.ids))
	}
}

func TestNotificationQueryMarkRead(t *testing.T) {
	id := domain.NotificationID("33333333-3333-3333-3333-333333333333")
	user := domain.UserID("22222222-2222-2222-2222-222222222222")
	inbox := &memInbox{items: []domain.InAppNotification{{
		ID: id, RecipientUserID: user, Title: "t", NotificationType: domain.TypeBonusCredited,
		Payload: json.RawMessage(`{}`), CreatedAt: time.Now().UTC(), OccurredAt: time.Now().UTC(),
	}}}
	q := application.NewNotificationQueryService(inbox)
	if err := q.MarkRead(context.Background(), string(user), string(id)); err != nil {
		t.Fatal(err)
	}
	if inbox.items[0].ReadAt == nil {
		t.Fatal("expected read")
	}
	unread := "unread"
	res, err := q.ListMine(context.Background(), string(user), application.ListQuery{
		Page: 1, PageSize: 20, ReadStatus: &unread,
	})
	if err != nil || len(res.Items) != 0 {
		t.Fatalf("unread list len=%d err=%v", len(res.Items), err)
	}
}
