package finalcheck

import (
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	finalcheckhttp "github.com/go-mentorship-platform/backend/internal/finalcheck/adapter/http"
	finalcheckpersistence "github.com/go-mentorship-platform/backend/internal/finalcheck/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/finalcheck/application"
)

// Module wires the final-check bounded context.
type Module struct {
	HTTP *finalcheckhttp.Module
}

// NewModule constructs final-check services and HTTP module.
func NewModule(pool *postgres.Pool, identityHTTP *identityhttp.Module) *Module {
	tx := finalcheckpersistence.NewTransactor(pool)
	repo := finalcheckpersistence.NewAssessmentRepo(pool)
	roadmap := finalcheckpersistence.NewRoadmapCompletionReader(pool)
	buddy := finalcheckpersistence.NewBuddyScopeReader(pool)
	events := finalcheckpersistence.NewOutboxRecorder(pool)

	svc := application.NewFinalCheckService(repo, roadmap, buddy, tx, events)
	query := application.NewFinalCheckQueryService(repo, roadmap, buddy, tx, events)

	httpModule := &finalcheckhttp.Module{
		Student: finalcheckhttp.NewStudentHandler(query),
		Buddy:   finalcheckhttp.NewBuddyHandler(svc, query),
		Admin:   finalcheckhttp.NewAdminHandler(svc, query),
		Authenticate:      identityHTTP.Authenticate,
		RequireActiveRole: identityHTTP.RequireActiveRole,
	}
	return &Module{HTTP: httpModule}
}
