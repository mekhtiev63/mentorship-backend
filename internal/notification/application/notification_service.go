package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-mentorship-platform/backend/internal/notification/domain"
)

// NotificationService ingests outbox events into the inbox.
type NotificationService struct {
	inbox   domain.InAppNotificationRepository
	outbox  domain.OutboxConsumerPort
	receipt domain.OutboxReceiptRepository
	lookup  domain.OneOnOneRequestLookup
	tx      domain.Transactor
	log     *slog.Logger
}

// NewNotificationService builds NotificationService.
func NewNotificationService(
	inbox domain.InAppNotificationRepository,
	outbox domain.OutboxConsumerPort,
	receipt domain.OutboxReceiptRepository,
	lookup domain.OneOnOneRequestLookup,
	tx domain.Transactor,
	log *slog.Logger,
) *NotificationService {
	return &NotificationService{
		inbox: inbox, outbox: outbox, receipt: receipt, lookup: lookup, tx: tx, log: log,
	}
}

// ProcessPending ingests a batch of outbox messages.
func (s *NotificationService) ProcessPending(ctx context.Context, limit int) (int, error) {
	msgs, err := s.outbox.ListUnprocessedForNotification(ctx, limit)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, msg := range msgs {
		if err := s.ingestOne(ctx, msg); err != nil {
			if s.log != nil {
				s.log.Error("notification ingest failed", "outboxId", msg.ID, "err", err)
			}
			continue
		}
		n++
	}
	return n, nil
}

func (s *NotificationService) ingestOne(ctx context.Context, msg domain.OutboxMessage) error {
	notifications, err := domain.MapOutboxMessage(ctx, msg, s.lookup)
	if err != nil {
		return err
	}
	if len(notifications) == 0 {
		return nil
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		for _, n := range notifications {
			if err := s.inbox.Append(ctx, n); err != nil && !errors.Is(err, domain.ErrDuplicate) {
				return err
			}
		}
		return s.receipt.InsertReceipt(ctx, msg.ID)
	})
}

// RunOutboxWorker polls outbox until ctx cancelled.
func (s *NotificationService) RunOutboxWorker(ctx context.Context) {
	if s.log != nil {
		s.log.Info("notification outbox worker started")
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
