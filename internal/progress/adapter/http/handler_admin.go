package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/progress/application"
)

// AdminHandler serves admin progress endpoints.
type AdminHandler struct {
	approval *application.BlockApprovalService
	query    *application.StudentProgressQueryService
}

// NewAdminHandler creates AdminHandler.
func NewAdminHandler(
	approval *application.BlockApprovalService,
	query *application.StudentProgressQueryService,
) *AdminHandler {
	return &AdminHandler{approval: approval, query: query}
}

// GetStudentProgress GET /admin/progress/students/{studentId}/blocks
func (h *AdminHandler) GetStudentProgress(w http.ResponseWriter, r *http.Request) {
	items, err := h.query.ListMyBlocks(r.Context(), chi.URLParam(r, "studentId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items})
}

// ApproveBlock POST /admin/progress/students/{studentId}/blocks/{blockId}/approve
func (h *AdminHandler) ApproveBlock(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	if err := h.approval.ApproveBlockAsAdmin(r.Context(), string(p.UserID), chi.URLParam(r, "studentId"), chi.URLParam(r, "blockId")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RejectBlock POST /admin/progress/students/{studentId}/blocks/{blockId}/reject
func (h *AdminHandler) RejectBlock(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var req rejectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	if err := h.approval.RejectBlockAsAdmin(r.Context(), string(p.UserID), chi.URLParam(r, "studentId"), chi.URLParam(r, "blockId"), req.Reason); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
