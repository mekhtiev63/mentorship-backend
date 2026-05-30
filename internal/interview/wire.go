package interview

import (
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	interviewhttp "github.com/go-mentorship-platform/backend/internal/interview/adapter/http"
	interviewpersistence "github.com/go-mentorship-platform/backend/internal/interview/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/interview/application"
)

// Module wires the interview bounded context.
type Module struct {
	HTTP *interviewhttp.Module
}

// NewModule constructs interview services and HTTP module.
func NewModule(pool *postgres.Pool, identityHTTP *identityhttp.Module) *Module {
	tx := interviewpersistence.NewTransactor(pool)
	repo := interviewpersistence.NewInterviewRepo(pool)
	buddy := interviewpersistence.NewBuddyScopeReader(pool)
	events := interviewpersistence.NewOutboxRecorder(pool)

	real := application.NewRealInterviewService(repo, tx, events)
	mock := application.NewMockInterviewService(repo, buddy, events)
	query := application.NewInterviewQueryService(repo)
	feedback := application.NewInterviewFeedbackService(repo, buddy, tx, events)

	httpModule := &interviewhttp.Module{
		Student: interviewhttp.NewStudentHandler(real, query, feedback),
		Buddy:   interviewhttp.NewBuddyHandler(mock, feedback),
		Authenticate:      identityHTTP.Authenticate,
		RequireActiveRole: identityHTTP.RequireActiveRole,
	}
	return &Module{HTTP: httpModule}
}
