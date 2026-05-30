package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
)

// OutboxRecorder writes domain events.
type OutboxRecorder struct {
	pool *postgres.Pool
}

// NewOutboxRecorder creates OutboxRecorder.
func NewOutboxRecorder(pool *postgres.Pool) *OutboxRecorder {
	return &OutboxRecorder{pool: pool}
}

// Record inserts an outbox row using the current transaction if present.
func (r *OutboxRecorder) Record(ctx context.Context, name string, payload map[string]any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal outbox: %w", err)
	}
	const q = `INSERT INTO outbox_events (event_name, payload) VALUES ($1, $2::jsonb)`
	_, err = querier(ctx, r.pool).Exec(ctx, q, name, raw)
	return err
}
