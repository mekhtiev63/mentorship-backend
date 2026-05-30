package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// Module wires interview HTTP routes.
type Module struct {
	Student *StudentHandler
	Buddy   *BuddyHandler

	Authenticate      func(http.Handler) http.Handler
	RequireActiveRole func(...identitydomain.Role) func(http.Handler) http.Handler
}

// Register mounts interview routes.
func (m *Module) Register(r chi.Router) {
	if m.Student != nil {
		r.Route("/interviews", func(r chi.Router) {
			r.Get("/catalog", m.Student.ListCatalog)

			r.Group(func(r chi.Router) {
				r.Use(m.Authenticate)
				r.Use(m.RequireActiveRole(identitydomain.RoleStudent))

				r.Route("/real", func(r chi.Router) {
					r.Post("/", m.Student.CreateReal)
					r.Get("/", m.Student.ListReal)
					r.Get("/{interviewId}", m.Student.GetReal)
					r.Patch("/{interviewId}", m.Student.UpdateReal)
					r.Post("/{interviewId}/complete", m.Student.CompleteReal)
				})
				r.Route("/mock", func(r chi.Router) {
					r.Get("/", m.Student.ListMock)
					r.Get("/{interviewId}", m.Student.GetMock)
					r.Get("/{interviewId}/feedback", m.Student.GetMockFeedback)
				})
			})
		})
	}
	if m.Buddy != nil && m.Authenticate != nil {
		r.Route("/buddy/interviews", func(r chi.Router) {
			r.Use(m.Authenticate)
			r.Use(m.RequireActiveRole(identitydomain.RoleBuddy))
			r.Route("/mock", func(r chi.Router) {
				r.Post("/", m.Buddy.CreateMock)
				r.Get("/", m.Buddy.ListMock)
				r.Post("/{interviewId}/complete", m.Buddy.CompleteMock)
				r.Get("/{interviewId}/feedback", m.Buddy.GetMockFeedback)
			})
		})
	}
}
