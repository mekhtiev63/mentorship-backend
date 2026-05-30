package achievement

import (
	"context"
	"log/slog"

	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	"github.com/go-mentorship-platform/backend/internal/platform/eventbus"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	achievementhttp "github.com/go-mentorship-platform/backend/internal/achievement/adapter/http"
	achievementpersistence "github.com/go-mentorship-platform/backend/internal/achievement/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/achievement/application"
)

// Module wires the achievement bounded context.
type Module struct {
	HTTP            *achievementhttp.Module
	progressService *application.AchievementProgressService
}

// NewModule constructs achievement services and HTTP module.
func NewModule(pool *postgres.Pool, identityHTTP *identityhttp.Module, bus *eventbus.Bus, log *slog.Logger) *Module {
	tx := achievementpersistence.NewTransactor(pool)
	definitions := achievementpersistence.NewDefinitionRepo(pool)
	grants := achievementpersistence.NewUserAchievementRepo(pool)
	progressStats := achievementpersistence.NewProgressStatsReader(pool)
	roadmapStats := achievementpersistence.NewRoadmapStatsReader(pool)
	outbox := achievementpersistence.NewOutboxRepo(pool)
	outboxWriter := achievementpersistence.NewOutboxRecorder(pool)
	buddy := achievementpersistence.NewBuddyScopeReader(pool)

	grantSvc := application.NewAchievementGrantService(tx, definitions, grants, progressStats, roadmapStats, outboxWriter)
	progressSvc := application.NewAchievementProgressService(outbox, grantSvc, tx, log)
	progressSvc.SubscribeBus(bus)

	catalogSvc := application.NewAchievementCatalogService(definitions, grants, buddy)

	httpModule := &achievementhttp.Module{
		Handler:      achievementhttp.NewHandler(catalogSvc),
		Authenticate: identityHTTP.Authenticate,
	}

	return &Module{
		HTTP:            httpModule,
		progressService: progressSvc,
	}
}

// RunOutboxWorker starts background outbox processing.
func (m *Module) RunOutboxWorker(ctx context.Context) {
	if m.progressService != nil {
		m.progressService.RunOutboxWorker(ctx)
	}
}
