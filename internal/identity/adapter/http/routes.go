package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	identitymw "github.com/go-mentorship-platform/backend/internal/identity/adapter/http/middleware"
	"github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// Module exposes HTTP registration and middleware for composition root.
type Module struct {
	Handler           *Handler
	Authenticate      func(http.Handler) http.Handler
	RequireRole       func(...domain.Role) func(http.Handler) http.Handler
	RequireActiveRole func(...domain.Role) func(http.Handler) http.Handler
}

// Register mounts auth routes on the router.
func (m *Module) Register(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", m.Handler.Login)
		r.Post("/logout", m.Handler.Logout)

		r.Group(func(r chi.Router) {
			r.Use(m.Authenticate)
			r.Get("/me", m.Handler.Me)
			r.Put("/active-role", m.Handler.SetActiveRole)
		})
	})
}

// NewModule wires handler and middleware.
func NewModule(handler *Handler, parser domain.TokenParser) *Module {
	return &Module{
		Handler:           handler,
		Authenticate:      identitymw.Authenticate(parser),
		RequireRole:       identitymw.RequireRole,
		RequireActiveRole: identitymw.RequireActiveRole,
	}
}
