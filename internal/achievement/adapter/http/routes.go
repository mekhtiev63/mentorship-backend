package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Module wires achievement HTTP routes.
type Module struct {
	Handler      *Handler
	Authenticate func(http.Handler) http.Handler
}

// Register mounts achievement routes.
func (m *Module) Register(r chi.Router) {
	r.Get("/achievements", m.Handler.ListCatalog)
	r.Group(func(r chi.Router) {
		r.Use(m.Authenticate)
		r.Get("/me/achievements", m.Handler.ListMyAchievements)
		r.Get("/users/{userId}/achievements", m.Handler.ListUserAchievements)
	})
}
