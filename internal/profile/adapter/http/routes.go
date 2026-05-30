package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Module wires profile routes.
type Module struct {
	Handler      *Handler
	Authenticate func(http.Handler) http.Handler
}

// Register mounts profile routes.
func (m *Module) Register(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(m.Authenticate)
		r.Get("/me/profile", m.Handler.GetMyProfile)
		r.Patch("/me/profile", m.Handler.PatchMyProfile)
		r.Get("/users/{userId}/profile", m.Handler.GetUserProfile)
	})
}
