package persistence

import (
	"context"
	"errors"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/notification/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// OutboxRepo reads outbox for notification ingestion.
type OutboxRepo struct {
	pool *postgres.Pool
}

// NewOutboxRepo creates OutboxRepo.
func NewOutboxRepo(pool *postgres.Pool) *OutboxRepo {
	return &OutboxRepo{pool: pool}
}

// ListUnprocessedForNotification returns whitelisted outbox rows without receipt.
func (r *OutboxRepo) ListUnprocessedForNotification(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	names := domain.NotificationSourceEventNames()
	const q = `
		SELECT o.id, o.event_name, o.payload, o.created_at
		FROM outbox_events o
		WHERE o.event_name = ANY($1)
		  AND NOT EXISTS (
			SELECT 1 FROM notification_outbox_receipts r WHERE r.outbox_id = o.id
		  )
		ORDER BY o.created_at ASC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, q, names, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.OutboxMessage
	for rows.Next() {
		var m domain.OutboxMessage
		if err := rows.Scan(&m.ID, &m.EventName, &m.Payload, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ReceiptRepo records processed outbox ids.
type ReceiptRepo struct {
	pool *postgres.Pool
}

// NewReceiptRepo creates ReceiptRepo.
func NewReceiptRepo(pool *postgres.Pool) *ReceiptRepo {
	return &ReceiptRepo{pool: pool}
}

// InsertReceipt marks outbox row processed for notification BC.
func (r *ReceiptRepo) InsertReceipt(ctx context.Context, outboxID string) error {
	const q = `INSERT INTO notification_outbox_receipts (outbox_id) VALUES ($1)`
	_, err := db(ctx, r.pool).Exec(ctx, q, outboxID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil
		}
		return err
	}
	return nil
}

// OneOnOneLookup resolves student id from request (ACL).
type OneOnOneLookup struct {
	pool *postgres.Pool
}

// NewOneOnOneLookup creates OneOnOneLookup.
func NewOneOnOneLookup(pool *postgres.Pool) *OneOnOneLookup {
	return &OneOnOneLookup{pool: pool}
}

// StudentIDByRequestID returns student_id for request.
func (l *OneOnOneLookup) StudentIDByRequestID(ctx context.Context, requestID string) (domain.UserID, error) {
	if _, err := uuid.Parse(requestID); err != nil {
		return "", domain.ErrValidation
	}
	const q = `SELECT student_id FROM one_on_one_requests WHERE id = $1`
	var sid string
	err := l.pool.QueryRow(ctx, q, requestID).Scan(&sid)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return domain.UserID(sid), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
