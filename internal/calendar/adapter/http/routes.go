package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Module wires calendar HTTP routes.
type Module struct {
	Handler      *Handler
	Authenticate func(http.Handler) http.Handler
}

// Register mounts calendar routes.
func (m *Module) Register(r chi.Router) {
	if m.Handler == nil || m.Authenticate == nil {
		return
	}
	r.Route("/calendar", func(r chi.Router) {
		r.Use(m.Authenticate)
		r.Get("/events/upcoming", m.Handler.ListUpcoming)
		r.Get("/events", m.Handler.ListEvents)
		r.Post("/events", m.Handler.CreateEvent)
		r.Get("/events/{eventId}", m.Handler.GetEvent)
		r.Patch("/events/{eventId}", m.Handler.UpdateEvent)
		r.Post("/events/{eventId}/cancel", m.Handler.CancelEvent)
		r.Delete("/events/{eventId}", m.Handler.DeleteEvent)
	})
}
