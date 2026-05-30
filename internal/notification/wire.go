package notification

import (
	"context"
	"log/slog"

	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	notificationhttp "github.com/go-mentorship-platform/backend/internal/notification/adapter/http"
	notificationpersistence "github.com/go-mentorship-platform/backend/internal/notification/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/notification/application"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
)

// Module wires the notification bounded context.
type Module struct {
	HTTP   *notificationhttp.Module
	ingest *application.NotificationService
}

// NewModule constructs notification services and HTTP module.
func NewModule(pool *postgres.Pool, identityHTTP *identityhttp.Module, log *slog.Logger) *Module {
	tx := notificationpersistence.NewTransactor(pool)
	inbox := notificationpersistence.NewInboxRepo(pool)
	outbox := notificationpersistence.NewOutboxRepo(pool)
	receipt := notificationpersistence.NewReceiptRepo(pool)
	lookup := notificationpersistence.NewOneOnOneLookup(pool)

	ingest := application.NewNotificationService(inbox, outbox, receipt, lookup, tx, log)
	query := application.NewNotificationQueryService(inbox)

	httpModule := &notificationhttp.Module{
		Handler:           notificationhttp.NewHandler(query),
		Authenticate:      identityHTTP.Authenticate,
		RequireActiveRole: identityHTTP.RequireActiveRole,
	}
	return &Module{HTTP: httpModule, ingest: ingest}
}

// RunOutboxWorker starts notification ingestion worker.
func (m *Module) RunOutboxWorker(ctx context.Context) {
	if m.ingest != nil {
		m.ingest.RunOutboxWorker(ctx)
	}
}
