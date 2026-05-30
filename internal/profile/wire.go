package profile

import (
	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	profilehttp "github.com/go-mentorship-platform/backend/internal/profile/adapter/http"
	profilepersistence "github.com/go-mentorship-platform/backend/internal/profile/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/profile/application"
	userpersistence "github.com/go-mentorship-platform/backend/internal/user/adapter/persistence"
)

// Module wires the profile bounded context.
type Module struct {
	HTTP *profilehttp.Module
}

// NewProfileRepo exposes profile repository for user module bootstrap.
func NewProfileRepo(pool *postgres.Pool) *profilepersistence.ProfileRepo {
	return profilepersistence.NewProfileRepo(pool)
}

// NewModule constructs profile services and HTTP module.
func NewModule(pool *postgres.Pool, identityHTTP *identityhttp.Module, assignments *userpersistence.BuddyAssignmentRepo) *Module {
	profiles := profilepersistence.NewProfileRepo(pool)
	users := profilepersistence.NewUserExistsRepo(pool)
	buddies := profilepersistence.NewBuddyReader(assignments)

	svc := application.NewProfileService(profiles, users, buddies)
	handler := profilehttp.NewHandler(svc)

	return &Module{
		HTTP: &profilehttp.Module{
			Handler:      handler,
			Authenticate: identityHTTP.Authenticate,
		},
	}
}
