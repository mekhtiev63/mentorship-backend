package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-mentorship-platform/backend/internal/activity/domain"
)

// ActivityService ingests outbox events into the activity journal.
type ActivityService struct {
	journal domain.ActivityJournalRepository
	outbox  domain.OutboxConsumerPort
	tx      domain.Transactor
	log     *slog.Logger
}

// NewActivityService builds ActivityService.
func NewActivityService(
	journal domain.ActivityJournalRepository,
	outbox domain.OutboxConsumerPort,
	tx domain.Transactor,
	log *slog.Logger,
) *ActivityService {
	return &ActivityService{journal: journal, outbox: outbox, tx: tx, log: log}
}

// ProcessPending ingests a batch of unprocessed outbox messages.
func (s *ActivityService) ProcessPending(ctx context.Context, limit int) (int, error) {
	msgs, err := s.outbox.ListUnprocessedForActivity(ctx, limit)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, msg := range msgs {
		if err := s.ingestOne(ctx, msg); err != nil {
			if s.log != nil {
				s.log.Error("activity ingest failed", "outboxId", msg.ID, "err", err)
			}
			continue
		}
		n++
	}
	return n, nil
}

func (s *ActivityService) ingestOne(ctx context.Context, msg domain.OutboxMessage) error {
	entry, ok := domain.MapOutboxMessage(msg)
	if !ok {
		return nil
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		err := s.journal.Append(ctx, entry)
		if errors.Is(err, domain.ErrDuplicate) {
			return nil
		}
		return err
	})
}

// RunOutboxWorker polls outbox until ctx cancelled.
func (s *ActivityService) RunOutboxWorker(ctx context.Context) {
	if s.log != nil {
		s.log.Info("activity outbox worker started")
	}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = s.ProcessPending(ctx, 50)
		}
	}
}
