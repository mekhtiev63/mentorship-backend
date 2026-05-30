package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// Module wires one-on-one HTTP routes.
type Module struct {
	Student *StudentHandler
	Admin   *AdminHandler

	Authenticate      func(http.Handler) http.Handler
	RequireActiveRole func(...identitydomain.Role) func(http.Handler) http.Handler
}

// Register mounts student routes on the API router.
func (m *Module) Register(r chi.Router) {
	if m.Student == nil {
		return
	}
	r.Route("/one-on-one", func(r chi.Router) {
		r.Use(m.Authenticate)
		r.Use(m.RequireActiveRole(identitydomain.RoleStudent))
		r.Post("/requests", m.Student.CreateRequest)
		r.Get("/requests", m.Student.ListRequests)
		r.Get("/requests/{requestId}", m.Student.GetRequest)
		r.Post("/requests/{requestId}/cancel", m.Student.CancelRequest)
	})
}

// RegisterAdmin mounts admin routes (caller applies admin middleware).
func (m *Module) RegisterAdmin(r chi.Router) {
	if m.Admin == nil {
		return
	}
	r.Route("/one-on-one", func(r chi.Router) {
		r.Get("/requests", m.Admin.ListRequests)
		r.Get("/requests/{requestId}", m.Admin.GetRequest)
		r.Post("/requests/{requestId}/approve", m.Admin.ApproveRequest)
		r.Post("/requests/{requestId}/reject", m.Admin.RejectRequest)
	})
}
