package bonus

import (
	"context"
	"log/slog"

	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	bonushttp "github.com/go-mentorship-platform/backend/internal/bonus/adapter/http"
	bonuspersistence "github.com/go-mentorship-platform/backend/internal/bonus/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/bonus/application"
)

// Module wires the bonus bounded context.
type Module struct {
	HTTP     *bonushttp.Module
	Ledger   *application.BonusLedgerService
	listener *application.BonusAchievementListenerService
}

// NewModule constructs bonus services and HTTP module.
func NewModule(pool *postgres.Pool, identityHTTP *identityhttp.Module, log *slog.Logger) *Module {
	tx := bonuspersistence.NewTransactor(pool)
	accounts := bonuspersistence.NewAccountRepo(pool)
	ledger := bonuspersistence.NewTransactionRepo(pool)
	outbox := bonuspersistence.NewOutboxRepo(pool)
	events := bonuspersistence.NewOutboxRecorder(pool)

	ledgerSvc := application.NewBonusLedgerService(tx, accounts, ledger, events)
	convertSvc := application.NewDiscountConversionService(ledgerSvc, accounts, ledger)
	balanceSvc := application.NewBonusBalanceQueryService(accounts, ledger)
	listener := application.NewBonusAchievementListenerService(outbox, ledgerSvc, tx, log)

	httpModule := &bonushttp.Module{
		Handler:           bonushttp.NewHandler(balanceSvc, convertSvc),
		Authenticate:      identityHTTP.Authenticate,
		RequireActiveRole: identityHTTP.RequireActiveRole,
	}

	return &Module{
		HTTP:     httpModule,
		Ledger:   ledgerSvc,
		listener: listener,
	}
}

// RunOutboxWorker starts achievement.granted consumer.
func (m *Module) RunOutboxWorker(ctx context.Context) {
	if m.listener != nil {
		m.listener.RunOutboxWorker(ctx)
	}
}
