package user

import (
	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	userhttp "github.com/go-mentorship-platform/backend/internal/user/adapter/http"
	userpersistence "github.com/go-mentorship-platform/backend/internal/user/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/user/application"
)

// Module wires the user bounded context.
type Module struct {
	HTTP         *userhttp.Module
	Assignments  *userpersistence.BuddyAssignmentRepo
}

// NewModule constructs user services and HTTP module.
func NewModule(pool *postgres.Pool, identityHTTP *identityhttp.Module) *Module {
	users := userpersistence.NewUserRepo(pool)
	roles := userpersistence.NewRoleRepo(pool)
	assignments := userpersistence.NewBuddyAssignmentRepo(pool)
	hasher := userpersistence.NewPasswordHasher()
	bootstrap := userpersistence.NewProfileBootstrap(pool)
	sessions := userpersistence.NewSessionRevoker(pool)

	admin := application.NewAdminService(users, roles, hasher, bootstrap, sessions, assignments)
	buddy := application.NewBuddyService(assignments)

	httpModule := &userhttp.Module{
		Admin:             userhttp.NewAdminHandler(admin),
		Buddy:             userhttp.NewBuddyHandler(buddy),
		Authenticate:      identityHTTP.Authenticate,
		RequireRole:       identityHTTP.RequireRole,
		RequireActiveRole: identityHTTP.RequireActiveRole,
	}

	return &Module{
		HTTP:        httpModule,
		Assignments: assignments,
	}
}
