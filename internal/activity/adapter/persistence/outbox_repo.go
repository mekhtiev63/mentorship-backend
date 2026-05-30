package persistence

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/activity/domain"
)

// OutboxRepo reads outbox for activity ingestion.
type OutboxRepo struct {
	pool *postgres.Pool
}

// NewOutboxRepo creates OutboxRepo.
func NewOutboxRepo(pool *postgres.Pool) *OutboxRepo {
	return &OutboxRepo{pool: pool}
}

// ListUnprocessedForActivity returns outbox rows not yet projected to activity_events.
func (r *OutboxRepo) ListUnprocessedForActivity(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	names := domain.ActivitySourceEventNames()
	const q = `
		SELECT o.id, o.event_name, o.payload, o.created_at
		FROM outbox_events o
		WHERE o.event_name = ANY($1)
		  AND NOT EXISTS (
			SELECT 1 FROM activity_events a WHERE a.source_outbox_id = o.id
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
