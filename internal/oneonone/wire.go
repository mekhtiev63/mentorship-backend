package oneonone

import (
	bonusapp "github.com/go-mentorship-platform/backend/internal/bonus/application"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	oneononebonus "github.com/go-mentorship-platform/backend/internal/oneonone/adapter/bonus"
	oneononehttp "github.com/go-mentorship-platform/backend/internal/oneonone/adapter/http"
	oneononepersistence "github.com/go-mentorship-platform/backend/internal/oneonone/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/oneonone/application"
	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
)

// Module wires the one-on-one bounded context.
type Module struct {
	HTTP *oneononehttp.Module
}

// NewModule constructs one-on-one services and HTTP module.
func NewModule(pool *postgres.Pool, identityHTTP *identityhttp.Module, ledger *bonusapp.BonusLedgerService) *Module {
	tx := oneononepersistence.NewTransactor(pool)
	requests := oneononepersistence.NewRequestRepo(pool)
	buddy := oneononepersistence.NewBuddyReader(pool)
	events := oneononepersistence.NewOutboxRecorder(pool)
	bonusPort := oneononebonus.NewLedgerAdapter(ledger)

	cmd := application.NewOneOnOneService(requests, buddy, bonusPort, events)
	admin := application.NewOneOnOneAdminService(requests, bonusPort, tx, events)
	query := application.NewOneOnOneQueryService(requests)

	httpModule := &oneononehttp.Module{
		Student: oneononehttp.NewStudentHandler(cmd, query),
		Admin:   oneononehttp.NewAdminHandler(admin, query),
		Authenticate:      identityHTTP.Authenticate,
		RequireActiveRole: identityHTTP.RequireActiveRole,
	}

	return &Module{HTTP: httpModule}
}
