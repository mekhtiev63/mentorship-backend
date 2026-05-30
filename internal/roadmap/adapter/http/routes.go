package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// Module wires roadmap HTTP routes.
type Module struct {
	Admin   *AdminHandler
	Catalog *CatalogHandler
	Student *StudentHandler

	Authenticate      func(http.Handler) http.Handler
	RequireRole       func(...identitydomain.Role) func(http.Handler) http.Handler
	RequireActiveRole func(...identitydomain.Role) func(http.Handler) http.Handler
}

// Register mounts public catalog and student routes on API v1.
func (m *Module) Register(r chi.Router) {
	if m.Catalog != nil {
		r.Route("/roadmap", func(r chi.Router) {
			r.Get("/blocks", m.Catalog.ListBlocks)
			r.Get("/blocks/{blockId}", m.Catalog.GetBlock)
		})
	}
	if m.Student != nil && m.Authenticate != nil {
		r.Route("/student", func(r chi.Router) {
			r.Use(m.Authenticate)
			r.Use(m.RequireActiveRole(identitydomain.RoleStudent))
			r.Get("/roadmap", m.Student.GetRoadmap)
			r.Get("/roadmap/blocks/{blockId}", m.Student.GetBlock)
			r.Get("/roadmap/blocks/{blockId}/materials", m.Student.ListMaterials)
		})
	}
}

// RegisterAdmin mounts admin roadmap routes (caller wraps /admin with auth).
func (m *Module) RegisterAdmin(r chi.Router) {
	if m.Admin == nil {
		return
	}
	r.Route("/roadmap", func(r chi.Router) {
		r.Get("/blocks", m.Admin.ListBlocks)
		r.Post("/blocks", m.Admin.CreateBlock)
		r.Put("/blocks/reorder", m.Admin.ReorderBlocks)
		r.Get("/blocks/{blockId}", m.Admin.GetBlock)
		r.Patch("/blocks/{blockId}", m.Admin.UpdateBlock)
		r.Delete("/blocks/{blockId}", m.Admin.DeleteBlock)
		r.Patch("/blocks/{blockId}/active", m.Admin.SetBlockActive)
		r.Post("/blocks/{blockId}/publish", m.Admin.PublishBlock)
		r.Post("/blocks/{blockId}/unpublish", m.Admin.UnpublishBlock)

		r.Get("/blocks/{blockId}/materials", m.Admin.ListMaterials)
		r.Post("/blocks/{blockId}/materials", m.Admin.CreateMaterial)
		r.Put("/blocks/{blockId}/materials/reorder", m.Admin.ReorderMaterials)

		r.Patch("/materials/{materialId}", m.Admin.UpdateMaterial)
		r.Delete("/materials/{materialId}", m.Admin.DeleteMaterial)
		r.Patch("/materials/{materialId}/active", m.Admin.SetMaterialActive)
	})
}
