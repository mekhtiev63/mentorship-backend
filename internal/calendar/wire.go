package calendar

import (
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	calendarhttp "github.com/go-mentorship-platform/backend/internal/calendar/adapter/http"
	calendarpersistence "github.com/go-mentorship-platform/backend/internal/calendar/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/calendar/application"
)

// Module wires the calendar bounded context.
type Module struct {
	HTTP *calendarhttp.Module
}

// NewModule constructs calendar services and HTTP module.
func NewModule(pool *postgres.Pool, identityHTTP *identityhttp.Module) *Module {
	tx := calendarpersistence.NewTransactor(pool)
	repo := calendarpersistence.NewEventRepo(pool)
	events := application.NewCalendarEventService(repo, tx)
	query := application.NewCalendarQueryService(repo)

	httpModule := &calendarhttp.Module{
		Handler:      calendarhttp.NewHandler(events, query),
		Authenticate: identityHTTP.Authenticate,
	}
	return &Module{HTTP: httpModule}
}
