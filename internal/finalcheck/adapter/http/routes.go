package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// Module wires final-check HTTP routes.
type Module struct {
	Student *StudentHandler
	Buddy   *BuddyHandler
	Admin   *AdminHandler

	Authenticate      func(http.Handler) http.Handler
	RequireActiveRole func(...identitydomain.Role) func(http.Handler) http.Handler
}

// Register mounts student and buddy routes.
func (m *Module) Register(r chi.Router) {
	if m.Student != nil && m.Authenticate != nil {
		r.Route("/final-check", func(r chi.Router) {
			r.Use(m.Authenticate)
			r.Use(m.RequireActiveRole(identitydomain.RoleStudent))
			r.Get("/me", m.Student.GetMe)
		})
	}
	if m.Buddy != nil && m.Authenticate != nil {
		r.Route("/buddy/final-check", func(r chi.Router) {
			r.Use(m.Authenticate)
			r.Use(m.RequireActiveRole(identitydomain.RoleBuddy))
			r.Get("/students/{studentId}", m.Buddy.GetStudent)
			r.Route("/students/{studentId}/tech", func(r chi.Router) {
				r.Post("/schedule", m.Buddy.ScheduleTech)
				r.Post("/complete", m.Buddy.CompleteTech)
				r.Post("/fail", m.Buddy.FailTech)
			})
			r.Route("/students/{studentId}/roast", func(r chi.Router) {
				r.Post("/schedule", m.Buddy.ScheduleRoast)
				r.Post("/complete", m.Buddy.CompleteRoast)
				r.Post("/fail", m.Buddy.FailRoast)
			})
		})
	}
}

// RegisterAdmin mounts admin routes (caller applies admin middleware).
func (m *Module) RegisterAdmin(r chi.Router) {
	if m.Admin == nil {
		return
	}
	r.Route("/final-check", func(r chi.Router) {
		r.Get("/students/{studentId}", m.Admin.GetStudent)
		r.Route("/students/{studentId}/tech", func(r chi.Router) {
			r.Post("/schedule", m.Admin.ScheduleTech)
			r.Post("/complete", m.Admin.CompleteTech)
			r.Post("/fail", m.Admin.FailTech)
		})
		r.Route("/students/{studentId}/roast", func(r chi.Router) {
			r.Post("/schedule", m.Admin.ScheduleRoast)
			r.Post("/complete", m.Admin.CompleteRoast)
			r.Post("/fail", m.Admin.FailRoast)
		})
	})
}
