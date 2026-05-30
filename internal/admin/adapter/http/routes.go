package http

import (
	"github.com/go-chi/chi/v5"
	activityhttp "github.com/go-mentorship-platform/backend/internal/activity/adapter/http"
	finalcheckhttp "github.com/go-mentorship-platform/backend/internal/finalcheck/adapter/http"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
	oneononehttp "github.com/go-mentorship-platform/backend/internal/oneonone/adapter/http"
	"github.com/go-mentorship-platform/backend/internal/platform/httpserver/stub"
	progresshttp "github.com/go-mentorship-platform/backend/internal/progress/adapter/http"
	roadmaphttp "github.com/go-mentorship-platform/backend/internal/roadmap/adapter/http"
	userhttp "github.com/go-mentorship-platform/backend/internal/user/adapter/http"
)

// RegisterDeps bundles handlers mounted under /admin.
type RegisterDeps struct {
	User       *userhttp.Module
	Roadmap    *roadmaphttp.Module
	Progress   *progresshttp.Module
	OneOnOne   *oneononehttp.Module
	FinalCheck *finalcheckhttp.Module
	Activity   *activityhttp.Module
}

// Register mounts admin API routes.
func Register(r chi.Router, deps RegisterDeps) {
	r.Route("/admin", func(r chi.Router) {
		if deps.User != nil || deps.Roadmap != nil || deps.Progress != nil {
			r.Group(func(r chi.Router) {
				switch {
				case deps.User != nil:
					r.Use(deps.User.Authenticate)
					r.Use(deps.User.RequireRole(identitydomain.RoleAdmin))
					r.Use(deps.User.RequireActiveRole(identitydomain.RoleAdmin))
				case deps.Roadmap != nil:
					r.Use(deps.Roadmap.Authenticate)
					r.Use(deps.Roadmap.RequireRole(identitydomain.RoleAdmin))
					r.Use(deps.Roadmap.RequireActiveRole(identitydomain.RoleAdmin))
				}
				if deps.User != nil {
					deps.User.RegisterAdmin(r)
				}
				if deps.Roadmap != nil {
					deps.Roadmap.RegisterAdmin(r)
				}
				if deps.Progress != nil {
					deps.Progress.RegisterAdmin(r)
				}
				if deps.OneOnOne != nil {
					deps.OneOnOne.RegisterAdmin(r)
				}
				if deps.FinalCheck != nil {
					deps.FinalCheck.RegisterAdmin(r)
				}
				if deps.Activity != nil {
					deps.Activity.RegisterAdmin(r)
				}
			})
		}
		r.Route("/achievements", func(r chi.Router) {
			r.Post("/", stub.Handler("admin"))
		})
	})
}
