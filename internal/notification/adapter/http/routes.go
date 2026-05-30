package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// Module wires notification HTTP routes.
type Module struct {
	Handler           *Handler
	Authenticate      func(http.Handler) http.Handler
	RequireActiveRole func(...identitydomain.Role) func(http.Handler) http.Handler
}

// Register mounts notification routes for authenticated platform roles.
func (m *Module) Register(r chi.Router) {
	if m.Handler == nil || m.Authenticate == nil {
		return
	}
	r.Route("/notifications", func(r chi.Router) {
		r.Use(m.Authenticate)
		r.Use(m.RequireActiveRole(
			identitydomain.RoleStudent,
			identitydomain.RoleBuddy,
			identitydomain.RoleAdmin,
		))
		r.Get("/", m.Handler.List)
		r.Get("/unread-count", m.Handler.UnreadCount)
		r.Post("/read-all", m.Handler.MarkAllRead)
		r.Patch("/{notificationId}/read", m.Handler.MarkRead)
	})
}
