package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// Module wires user HTTP routes.
type Module struct {
	Admin *AdminHandler
	Buddy *BuddyHandler

	Authenticate     func(http.Handler) http.Handler
	RequireRole      func(...identitydomain.Role) func(http.Handler) http.Handler
	RequireActiveRole func(...identitydomain.Role) func(http.Handler) http.Handler

	// ExtraBuddyRegister mounts additional routes under /buddy (same auth middleware).
	ExtraBuddyRegister func(chi.Router)
}

// RegisterAdmin mounts admin routes (caller must wrap /admin with auth).
func (m *Module) RegisterAdmin(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.Get("/", m.Admin.ListUsers)
		r.Post("/", m.Admin.CreateUser)
		r.Get("/{userId}", m.Admin.GetUser)
		r.Patch("/{userId}/status", m.Admin.UpdateStatus)
		r.Put("/{userId}/roles", m.Admin.ReplaceRoles)
		r.Delete("/{userId}", m.Admin.DeleteUser)
	})
	r.Route("/buddy-assignments", func(r chi.Router) {
		r.Post("/", m.Admin.CreateBuddyAssignment)
		r.Delete("/{assignmentId}", m.Admin.DeleteBuddyAssignment)
	})
}

// Register mounts buddy routes under /buddy.
func (m *Module) Register(r chi.Router) {
	r.Route("/buddy", func(r chi.Router) {
		r.Use(m.Authenticate)
		r.Use(m.RequireActiveRole(identitydomain.RoleBuddy))
		r.Get("/students", m.Buddy.ListStudents)
		r.Get("/students/{studentId}", m.Buddy.GetStudent)
		if m.ExtraBuddyRegister != nil {
			m.ExtraBuddyRegister(r)
		}
	})
}
