package persistence

import (
	"context"
	"encoding/json"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/achievement/domain"
)

// OutboxRepo reads progress-related outbox events.
type OutboxRepo struct {
	pool *postgres.Pool
}

// NewOutboxRepo creates OutboxRepo.
func NewOutboxRepo(pool *postgres.Pool) *OutboxRepo {
	return &OutboxRepo{pool: pool}
}

// ListPendingProgressEvents returns pending progress.* events.
func (r *OutboxRepo) ListPendingProgressEvents(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	const q = `
		SELECT id, event_name, payload, created_at
		FROM outbox_events
		WHERE status = 'pending' AND event_name LIKE 'progress.%'
		ORDER BY created_at ASC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, q, limit)
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

// MarkDone marks outbox row processed.
func (r *OutboxRepo) MarkDone(ctx context.Context, id string) error {
	const q = `
		UPDATE outbox_events
		SET status = 'done', processed_at = now()
		WHERE id = $1 AND status = 'pending'
	`
	_, err := db(ctx, r.pool).Exec(ctx, q, id)
	return err
}

// OutboxRecorder records new events.
type OutboxRecorder struct {
	pool *postgres.Pool
}

// NewOutboxRecorder creates OutboxRecorder.
func NewOutboxRecorder(pool *postgres.Pool) *OutboxRecorder {
	return &OutboxRecorder{pool: pool}
}

// Record inserts outbox row.
func (r *OutboxRecorder) Record(ctx context.Context, name string, payload map[string]any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	const q = `INSERT INTO outbox_events (event_name, payload) VALUES ($1, $2::jsonb)`
	_, err = db(ctx, r.pool).Exec(ctx, q, name, raw)
	return err
}
