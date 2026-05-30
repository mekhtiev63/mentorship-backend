package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// Module wires progress HTTP routes.
type Module struct {
	Student *StudentHandler
	Buddy   *BuddyHandler
	Admin   *AdminHandler

	Authenticate      func(http.Handler) http.Handler
	RequireRole       func(...identitydomain.Role) func(http.Handler) http.Handler
	RequireActiveRole func(...identitydomain.Role) func(http.Handler) http.Handler
}

// Register mounts student and buddy progress routes.
func (m *Module) Register(r chi.Router) {
	if m.Student != nil {
		r.Route("/progress", func(r chi.Router) {
			r.Use(m.Authenticate)
			r.Use(m.RequireActiveRole(identitydomain.RoleStudent))
			r.Get("/blocks", m.Student.ListBlocks)
			r.Get("/blocks/{blockId}", m.Student.GetBlock)
			r.Post("/materials/{materialId}/view", m.Student.RecordMaterialView)
			r.Post("/blocks/{blockId}/submit", m.Student.SubmitBlock)
		})
	}
	if m.Buddy != nil {
		r.Route("/buddy/progress", func(r chi.Router) {
			r.Use(m.Authenticate)
			r.Use(m.RequireActiveRole(identitydomain.RoleBuddy))
			r.Get("/students", m.Buddy.ListStudents)
			r.Get("/students/{studentId}", m.Buddy.GetStudentProgress)
			r.Get("/approvals", m.Buddy.ListApprovals)
			r.Post("/students/{studentId}/blocks/{blockId}/approve", m.Buddy.ApproveBlock)
			r.Post("/students/{studentId}/blocks/{blockId}/reject", m.Buddy.RejectBlock)
		})
	}
}

// RegisterAdmin mounts admin progress routes (caller applies admin middleware).
func (m *Module) RegisterAdmin(r chi.Router) {
	if m.Admin == nil {
		return
	}
	r.Route("/progress", func(r chi.Router) {
		r.Get("/students/{studentId}/blocks", m.Admin.GetStudentProgress)
		r.Post("/students/{studentId}/blocks/{blockId}/approve", m.Admin.ApproveBlock)
		r.Post("/students/{studentId}/blocks/{blockId}/reject", m.Admin.RejectBlock)
	})
}
