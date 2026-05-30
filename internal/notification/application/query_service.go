package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/notification/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// NotificationQueryService reads and updates inbox state for the current user.
type NotificationQueryService struct {
	inbox domain.InAppNotificationRepository
}

// NewNotificationQueryService builds NotificationQueryService.
func NewNotificationQueryService(inbox domain.InAppNotificationRepository) *NotificationQueryService {
	return &NotificationQueryService{inbox: inbox}
}

// ListMine returns paginated notifications for recipient.
func (s *NotificationQueryService) ListMine(ctx context.Context, userID string, q ListQuery) (ListResult, error) {
	uid, err := domain.ParseUserID(userID)
	if err != nil {
		return ListResult{}, err
	}
	p := pagination.Normalize(q.Page, q.PageSize)
	filter := toFilter(q)
	items, total, err := s.inbox.ListForRecipient(ctx, uid, filter, p)
	if err != nil {
		return ListResult{}, err
	}
	unread, err := s.inbox.CountUnread(ctx, uid)
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{
		Items:       toDTOs(items),
		Meta:        pagination.NewMeta(p.Page, p.PageSize, total),
		UnreadCount: unread,
	}, nil
}

// GetUnreadCount returns unread total for user.
func (s *NotificationQueryService) GetUnreadCount(ctx context.Context, userID string) (UnreadCountDTO, error) {
	uid, err := domain.ParseUserID(userID)
	if err != nil {
		return UnreadCountDTO{}, err
	}
	n, err := s.inbox.CountUnread(ctx, uid)
	if err != nil {
		return UnreadCountDTO{}, err
	}
	return UnreadCountDTO{Count: n}, nil
}

// MarkRead marks one notification read for owner.
func (s *NotificationQueryService) MarkRead(ctx context.Context, userID, notificationID string) error {
	uid, err := domain.ParseUserID(userID)
	if err != nil {
		return err
	}
	nid, err := domain.ParseNotificationID(notificationID)
	if err != nil {
		return domain.ErrNotFound
	}
	if _, err := s.inbox.GetByIDForRecipient(ctx, nid, uid); err != nil {
		return err
	}
	return s.inbox.MarkRead(ctx, nid, uid, time.Now().UTC())
}

// MarkAllRead marks all unread notifications read for user.
func (s *NotificationQueryService) MarkAllRead(ctx context.Context, userID string) (MarkAllReadResult, error) {
	uid, err := domain.ParseUserID(userID)
	if err != nil {
		return MarkAllReadResult{}, err
	}
	n, err := s.inbox.MarkAllRead(ctx, uid, time.Now().UTC())
	if err != nil {
		return MarkAllReadResult{}, err
	}
	return MarkAllReadResult{Updated: n}, nil
}

func toFilter(q ListQuery) domain.NotificationListFilter {
	var f domain.NotificationListFilter
	if q.ReadStatus != nil {
		switch *q.ReadStatus {
		case string(domain.ReadStatusUnread):
			rs := domain.ReadStatusUnread
			f.ReadStatus = &rs
		case string(domain.ReadStatusRead):
			rs := domain.ReadStatusRead
			f.ReadStatus = &rs
		}
	}
	if q.NotificationType != nil && *q.NotificationType != "" {
		t := domain.NotificationType(*q.NotificationType)
		f.NotificationType = &t
	}
	return f
}

func toDTOs(items []domain.InAppNotification) []NotificationDTO {
	out := make([]NotificationDTO, 0, len(items))
	for _, n := range items {
		out = append(out, toDTO(n))
	}
	return out
}

func toDTO(n domain.InAppNotification) NotificationDTO {
	var actor *string
	if n.ActorID != nil {
		s := string(*n.ActorID)
		actor = &s
	}
	return NotificationDTO{
		ID:               string(n.ID),
		NotificationType: string(n.NotificationType),
		Title:            n.Title,
		Body:             n.Body,
		Payload:          n.Payload,
		ActorID:          actor,
		ReferenceType:    n.ReferenceType,
		ReferenceID:      n.ReferenceID,
		OccurredAt:       n.OccurredAt,
		ReadAt:           n.ReadAt,
		CreatedAt:        n.CreatedAt,
		IsRead:           n.IsRead(),
	}
}
