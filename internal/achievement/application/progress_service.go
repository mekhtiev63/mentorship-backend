package application

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-mentorship-platform/backend/internal/achievement/domain"
	"github.com/go-mentorship-platform/backend/internal/platform/eventbus"
)

// AchievementProgressService consumes progress domain events (outbox + bus).
type AchievementProgressService struct {
	outbox domain.OutboxReader
	grant  *AchievementGrantService
	tx     domain.Transactor
	log    *slog.Logger
}

// NewAchievementProgressService builds AchievementProgressService.
func NewAchievementProgressService(
	outbox domain.OutboxReader,
	grant *AchievementGrantService,
	tx domain.Transactor,
	log *slog.Logger,
) *AchievementProgressService {
	return &AchievementProgressService{outbox: outbox, grant: grant, tx: tx, log: log}
}

// ProcessPendingOutbox processes a batch of pending progress outbox messages.
func (s *AchievementProgressService) ProcessPendingOutbox(ctx context.Context, limit int) (int, error) {
	msgs, err := s.outbox.ListPendingProgressEvents(ctx, limit)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, msg := range msgs {
		if err := s.handleOne(ctx, msg); err != nil {
			if s.log != nil {
				s.log.Error("achievement outbox handler failed", "id", msg.ID, "err", err)
			}
			continue
		}
		processed++
	}
	return processed, nil
}

func (s *AchievementProgressService) handleOne(ctx context.Context, msg domain.OutboxMessage) error {
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.grant.processEvent(ctx, msg.ID, msg.EventName, msg.Payload); err != nil {
			return err
		}
		return s.outbox.MarkDone(ctx, msg.ID)
	})
}

type busEvent struct {
	SourceID  string
	EventName string
	Payload   json.RawMessage
}

func (b busEvent) Name() string { return b.EventName }

// HandleBusEvent handles in-process bus delivery.
func (s *AchievementProgressService) HandleBusEvent(ctx context.Context, e eventbus.Event) error {
	be, ok := e.(busEvent)
	if !ok {
		return nil
	}
	return s.grant.HandleProgressEvent(ctx, be.SourceID, be.EventName, be.Payload)
}

// SubscribeBus registers achievement handlers on the event bus.
func (s *AchievementProgressService) SubscribeBus(bus *eventbus.Bus) {
	if bus == nil {
		return
	}
	handler := func(ctx context.Context, event eventbus.Event) error {
		return s.HandleBusEvent(ctx, event)
	}
	bus.Subscribe(domain.ProgressMaterialViewed, handler)
	bus.Subscribe(domain.ProgressBlockApproved, handler)
}

// RunOutboxWorker polls outbox until ctx is cancelled.
func (s *AchievementProgressService) RunOutboxWorker(ctx context.Context) {
	if s.log != nil {
		s.log.Info("achievement outbox worker started")
	}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = s.ProcessPendingOutbox(ctx, 50)
		}
	}
}
