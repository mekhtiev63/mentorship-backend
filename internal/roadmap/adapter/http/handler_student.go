package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/roadmap/application"
)

// StudentHandler serves authenticated student roadmap reads.
type StudentHandler struct {
	student *application.StudentRoadmapService
}

// NewStudentHandler creates StudentHandler.
func NewStudentHandler(student *application.StudentRoadmapService) *StudentHandler {
	return &StudentHandler{student: student}
}

// GetRoadmap GET /student/roadmap
func (h *StudentHandler) GetRoadmap(w http.ResponseWriter, r *http.Request) {
	roadmap, err := h.student.GetRoadmap(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, roadmap)
}

// GetBlock GET /student/roadmap/blocks/{blockId}
func (h *StudentHandler) GetBlock(w http.ResponseWriter, r *http.Request) {
	block, err := h.student.GetBlock(r.Context(), chi.URLParam(r, "blockId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, block)
}

// ListMaterials GET /student/roadmap/blocks/{blockId}/materials
func (h *StudentHandler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	items, err := h.student.ListMaterials(r.Context(), chi.URLParam(r, "blockId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items})
}
