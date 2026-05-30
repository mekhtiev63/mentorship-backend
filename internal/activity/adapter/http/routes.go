package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// Module wires activity HTTP routes.
type Module struct {
	Handler           *Handler
	Authenticate      func(http.Handler) http.Handler
	RequireActiveRole func(...identitydomain.Role) func(http.Handler) http.Handler
}

// Register mounts student activity routes.
func (m *Module) Register(r chi.Router) {
	if m.Handler == nil || m.Authenticate == nil {
		return
	}
	r.Route("/activity", func(r chi.Router) {
		r.Use(m.Authenticate)
		r.Use(m.RequireActiveRole(identitydomain.RoleStudent))
		r.Get("/me", m.Handler.ListMe)
	})
}

// RegisterBuddy mounts buddy activity routes.
func (m *Module) RegisterBuddy(r chi.Router) {
	if m.Handler == nil {
		return
	}
	r.Route("/activity", func(r chi.Router) {
		r.Get("/students/{studentId}", m.Handler.ListStudent)
	})
}

// RegisterAdmin mounts admin activity routes (caller applies admin middleware).
func (m *Module) RegisterAdmin(r chi.Router) {
	if m.Handler == nil {
		return
	}
	r.Route("/activity", func(r chi.Router) {
		r.Get("/", m.Handler.ListAdmin)
	})
}
