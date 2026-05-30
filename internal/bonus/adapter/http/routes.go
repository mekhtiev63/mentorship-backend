package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// Module wires bonus HTTP routes.
type Module struct {
	Handler           *Handler
	Authenticate      func(http.Handler) http.Handler
	RequireActiveRole func(...identitydomain.Role) func(http.Handler) http.Handler
}

// Register mounts bonus routes.
func (m *Module) Register(r chi.Router) {
	r.Route("/me/bonus", func(r chi.Router) {
		r.Use(m.Authenticate)
		r.Use(m.RequireActiveRole(identitydomain.RoleStudent))
		r.Get("/", m.Handler.GetBonus)
		r.Get("/transactions", m.Handler.ListTransactions)
		r.Post("/convert", m.Handler.Convert)
	})
}
