package identity

import (
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/identity/adapter/clock"
	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/security"
	"github.com/go-mentorship-platform/backend/internal/identity/application"
	"github.com/go-mentorship-platform/backend/internal/platform/config"
	platformpostgres "github.com/go-mentorship-platform/backend/internal/platform/postgres"
)

// Module is the wired identity bounded context.
type Module struct {
	HTTP *identityhttp.Module
}

// NewModule constructs identity application and HTTP adapters.
func NewModule(cfg config.Auth, pool *platformpostgres.Pool) (*Module, error) {
	secret := cfg.JWTSecret
	if secret == "" {
		secret = "development-only-change-me"
	}

	credentialsRepo := persistence.NewCredentialsRepo(pool)
	roleRepo := persistence.NewRoleRepo(pool)
	prefsRepo := persistence.NewPreferencesRepo(pool)
	refreshRepo := persistence.NewRefreshTokenRepo(pool)

	hasher := security.NewBcryptHasher()
	jwtProvider := security.NewJWTProvider(secret)
	refreshGen := security.NewRefreshGenerator()
	sysClock := clock.SystemClock{}

	loginSvc := application.NewLoginService(
		credentialsRepo,
		roleRepo,
		prefsRepo,
		refreshRepo,
		hasher,
		jwtProvider,
		refreshGen,
		sysClock,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)
	logoutSvc := application.NewLogoutService(refreshRepo)
	meSvc := application.NewMeService(credentialsRepo, roleRepo, prefsRepo)
	activeRoleSvc := application.NewActiveRoleService(
		credentialsRepo,
		roleRepo,
		prefsRepo,
		refreshRepo,
		jwtProvider,
		refreshGen,
		sysClock,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	handler := identityhttp.NewHandler(loginSvc, logoutSvc, meSvc, activeRoleSvc)
	httpModule := identityhttp.NewModule(handler, jwtProvider)

	return &Module{HTTP: httpModule}, nil
}

// MustNewModule calls NewModule and panics on error (for tests).
func MustNewModule(cfg config.Auth, pool *platformpostgres.Pool) *Module {
	m, err := NewModule(cfg, pool)
	if err != nil {
		panic(fmt.Sprintf("identity module: %v", err))
	}
	return m
}
