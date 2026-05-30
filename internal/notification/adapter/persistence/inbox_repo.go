package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/notification/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// InboxRepo implements InAppNotificationRepository.
type InboxRepo struct {
	pool *postgres.Pool
}

// NewInboxRepo creates InboxRepo.
func NewInboxRepo(pool *postgres.Pool) *InboxRepo {
	return &InboxRepo{pool: pool}
}

// Append inserts notification (idempotent by source_outbox_id + recipient + type).
func (r *InboxRepo) Append(ctx context.Context, n domain.InAppNotification) error {
	id := string(n.ID)
	if id == "" {
		id = uuid.NewString()
	}
	var actor *string
	if n.ActorID != nil {
		s := string(*n.ActorID)
		actor = &s
	}
	const q = `
		INSERT INTO in_app_notifications (
			id, recipient_user_id, notification_type, title, body, payload,
			actor_id, reference_type, reference_id,
			source_outbox_id, source_event_name, occurred_at, read_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := db(ctx, r.pool).Exec(ctx, q,
		id, string(n.RecipientUserID), string(n.NotificationType), n.Title, n.Body, n.Payload,
		actor, n.ReferenceType, n.ReferenceID,
		n.SourceOutboxID, n.SourceEventName, n.OccurredAt, n.ReadAt, n.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrDuplicate
		}
		return err
	}
	return nil
}

// GetByIDForRecipient loads notification for owner.
func (r *InboxRepo) GetByIDForRecipient(ctx context.Context, id domain.NotificationID, recipient domain.UserID) (domain.InAppNotification, error) {
	const q = selectNotification + ` WHERE id = $1 AND recipient_user_id = $2`
	return scanNotification(r.pool.QueryRow(ctx, q, string(id), string(recipient)))
}

// ListForRecipient lists inbox ordered by created_at DESC.
func (r *InboxRepo) ListForRecipient(ctx context.Context, recipient domain.UserID, filter domain.NotificationListFilter, page pagination.Params) ([]domain.InAppNotification, int64, error) {
	where := `WHERE recipient_user_id = $1`
	args := []any{string(recipient)}
	argn := 2
	if filter.ReadStatus != nil {
		switch *filter.ReadStatus {
		case domain.ReadStatusUnread:
			where += ` AND read_at IS NULL`
		case domain.ReadStatusRead:
			where += ` AND read_at IS NOT NULL`
		}
	}
	if filter.NotificationType != nil {
		where += fmt.Sprintf(` AND notification_type = $%d`, argn)
		args = append(args, string(*filter.NotificationType))
		argn++
	}

	countQ := `SELECT COUNT(*) FROM in_app_notifications ` + where
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ := selectNotification + ` ` + where + fmt.Sprintf(`
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, argn, argn+1)
	args = append(args, page.PageSize, page.Offset)

	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.InAppNotification
	for rows.Next() {
		n, err := scanNotificationRow(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, n)
	}
	return out, total, rows.Err()
}

// CountUnread returns unread count for recipient.
func (r *InboxRepo) CountUnread(ctx context.Context, recipient domain.UserID) (int64, error) {
	const q = `SELECT COUNT(*) FROM in_app_notifications WHERE recipient_user_id = $1 AND read_at IS NULL`
	var n int64
	err := r.pool.QueryRow(ctx, q, string(recipient)).Scan(&n)
	return n, err
}

// MarkRead sets read_at when still unread.
func (r *InboxRepo) MarkRead(ctx context.Context, id domain.NotificationID, recipient domain.UserID, readAt time.Time) error {
	const q = `
		UPDATE in_app_notifications
		SET read_at = $3
		WHERE id = $1 AND recipient_user_id = $2 AND read_at IS NULL
	`
	tag, err := db(ctx, r.pool).Exec(ctx, q, string(id), string(recipient), readAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	return nil
}

// MarkAllRead marks all unread rows for recipient.
func (r *InboxRepo) MarkAllRead(ctx context.Context, recipient domain.UserID, readAt time.Time) (int64, error) {
	const q = `
		UPDATE in_app_notifications
		SET read_at = $2
		WHERE recipient_user_id = $1 AND read_at IS NULL
	`
	tag, err := db(ctx, r.pool).Exec(ctx, q, string(recipient), readAt)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

const selectNotification = `
	SELECT id, recipient_user_id, notification_type, title, body, payload,
	       actor_id, reference_type, reference_id,
	       source_outbox_id, source_event_name, occurred_at, read_at, created_at
	FROM in_app_notifications
`

func scanNotification(row pgx.Row) (domain.InAppNotification, error) {
	return scanNotificationRow(row)
}

func scanNotificationRow(row pgx.Row) (domain.InAppNotification, error) {
	var n domain.InAppNotification
	var id, recipient, typ, title, body string
	var actor, refType, refID *string
	err := row.Scan(
		&id, &recipient, &typ, &title, &body, &n.Payload,
		&actor, &refType, &refID,
		&n.SourceOutboxID, &n.SourceEventName, &n.OccurredAt, &n.ReadAt, &n.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.InAppNotification{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.InAppNotification{}, err
	}
	n.ID = domain.NotificationID(id)
	n.RecipientUserID = domain.UserID(recipient)
	n.NotificationType = domain.NotificationType(typ)
	n.Title = title
	n.Body = body
	n.ReferenceType = refType
	n.ReferenceID = refID
	if actor != nil {
		u := domain.UserID(*actor)
		n.ActorID = &u
	}
	return n, nil
}
