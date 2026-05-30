package activity

import (
	"context"
	"log/slog"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	activityhttp "github.com/go-mentorship-platform/backend/internal/activity/adapter/http"
	activitypersistence "github.com/go-mentorship-platform/backend/internal/activity/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/activity/application"
)

// Module wires the activity bounded context.
type Module struct {
	HTTP    *activityhttp.Module
	ingest  *application.ActivityService
}

// NewModule constructs activity services and HTTP module.
func NewModule(pool *postgres.Pool, identityHTTP *identityhttp.Module, log *slog.Logger) *Module {
	tx := activitypersistence.NewTransactor(pool)
	journal := activitypersistence.NewJournalRepo(pool)
	outbox := activitypersistence.NewOutboxRepo(pool)
	buddy := activitypersistence.NewBuddyScopeReader(pool)

	ingest := application.NewActivityService(journal, outbox, tx, log)
	query := application.NewActivityQueryService(journal, buddy)

	httpModule := &activityhttp.Module{
		Handler:           activityhttp.NewHandler(query),
		Authenticate:      identityHTTP.Authenticate,
		RequireActiveRole: identityHTTP.RequireActiveRole,
	}
	return &Module{HTTP: httpModule, ingest: ingest}
}

// RunOutboxWorker starts activity ingestion worker.
func (m *Module) RunOutboxWorker(ctx context.Context) {
	if m.ingest != nil {
		m.ingest.RunOutboxWorker(ctx)
	}
}
