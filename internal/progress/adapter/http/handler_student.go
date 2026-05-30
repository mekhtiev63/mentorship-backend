package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/progress/application"
)

// StudentHandler serves student progress endpoints.
type StudentHandler struct {
	material *application.MaterialProgressService
	query    *application.StudentProgressQueryService
}

// NewStudentHandler creates StudentHandler.
func NewStudentHandler(
	material *application.MaterialProgressService,
	query *application.StudentProgressQueryService,
) *StudentHandler {
	return &StudentHandler{material: material, query: query}
}

func studentIDFromRequest(r *http.Request) (string, bool) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		return "", false
	}
	return string(p.UserID), true
}

// ListBlocks GET /progress/blocks
func (h *StudentHandler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	sid, ok := studentIDFromRequest(r)
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	items, err := h.query.ListMyBlocks(r.Context(), sid)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items})
}

// GetBlock GET /progress/blocks/{blockId}
func (h *StudentHandler) GetBlock(w http.ResponseWriter, r *http.Request) {
	sid, ok := studentIDFromRequest(r)
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	block, err := h.query.GetMyBlock(r.Context(), sid, chi.URLParam(r, "blockId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, block)
}

// RecordMaterialView POST /progress/materials/{materialId}/view
func (h *StudentHandler) RecordMaterialView(w http.ResponseWriter, r *http.Request) {
	sid, ok := studentIDFromRequest(r)
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var key *string
	if k := r.Header.Get("Idempotency-Key"); k != "" {
		key = &k
	}
	res, err := h.material.RecordMaterialView(r.Context(), sid, chi.URLParam(r, "materialId"), key)
	if err != nil {
		writeError(w, err)
		return
	}
	status := http.StatusOK
	if res.Created {
		status = http.StatusCreated
	}
	writeData(w, status, res)
}

// SubmitBlock POST /progress/blocks/{blockId}/submit
func (h *StudentHandler) SubmitBlock(w http.ResponseWriter, r *http.Request) {
	sid, ok := studentIDFromRequest(r)
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	block, err := h.material.SubmitBlock(r.Context(), sid, chi.URLParam(r, "blockId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, block)
}
