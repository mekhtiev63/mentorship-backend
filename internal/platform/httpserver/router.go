package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-mentorship-platform/backend/internal/platform/config"
	"github.com/go-mentorship-platform/backend/internal/platform/httpserver/health"
	"github.com/go-mentorship-platform/backend/internal/platform/httpserver/response"
	"github.com/go-mentorship-platform/backend/internal/platform/httpserver/routes"
	platformmw "github.com/go-mentorship-platform/backend/internal/platform/middleware"
)

// Deps are HTTP-layer dependencies assembled by the application container.
type Deps struct {
	Config config.Config
	Logger *slog.Logger
	Health *health.Handler
	V1     routes.V1Deps
}

// NewRouter registers global middleware and routes.
func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RealIP)
	r.Use(platformmw.RequestID)
	r.Use(platformmw.Recover(deps.Logger))
	r.Use(platformmw.Logger(deps.Logger))
	r.Use(platformmw.JSON)

	if deps.Config.App.Env == "development" {
		r.Use(platformmw.CORSAllowAll())
	} else {
		r.Use(platformmw.CORS(deps.Config.CORS.AllowedOrigins))
	}

	r.Get("/health/live", deps.Health.Live)
	r.Get("/health/ready", deps.Health.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
			response.OK(w, map[string]string{"message": "pong"})
		})
		routes.RegisterV1(r, deps.V1)
	})

	return r
}
