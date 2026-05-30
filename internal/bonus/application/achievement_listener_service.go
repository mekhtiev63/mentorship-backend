package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-mentorship-platform/backend/internal/bonus/domain"
)

// BonusAchievementListenerService processes achievement.granted outbox events.
type BonusAchievementListenerService struct {
	outbox  domain.OutboxReader
	ledger  *BonusLedgerService
	tx      domain.Transactor
	log     *slog.Logger
}

// NewBonusAchievementListenerService builds listener.
func NewBonusAchievementListenerService(
	outbox domain.OutboxReader,
	ledger *BonusLedgerService,
	tx domain.Transactor,
	log *slog.Logger,
) *BonusAchievementListenerService {
	return &BonusAchievementListenerService{outbox: outbox, ledger: ledger, tx: tx, log: log}
}

// ProcessPending processes a batch of achievement.granted messages.
func (s *BonusAchievementListenerService) ProcessPending(ctx context.Context, limit int) (int, error) {
	msgs, err := s.outbox.ListPendingAchievementGranted(ctx, limit)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, msg := range msgs {
		if err := s.handleOne(ctx, msg); err != nil {
			if s.log != nil {
				s.log.Error("bonus achievement handler failed", "id", msg.ID, "err", err)
			}
			continue
		}
		n++
	}
	return n, nil
}

func (s *BonusAchievementListenerService) handleOne(ctx context.Context, msg domain.OutboxMessage) error {
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		p, err := parseAchievementGranted(msg.Payload)
		if err != nil {
			return err
		}
		sourceID := p.SourceEventID
		if sourceID == "" {
			sourceID = msg.ID
		}
		_, err = s.ledger.creditAchievementInTx(ctx, p.UserID, p.AchievementCode, sourceID)
		if err != nil {
			return err
		}
		return s.outbox.MarkDone(ctx, msg.ID)
	})
}

// RunOutboxWorker polls until ctx cancelled.
func (s *BonusAchievementListenerService) RunOutboxWorker(ctx context.Context) {
	if s.log != nil {
		s.log.Info("bonus achievement outbox worker started")
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
