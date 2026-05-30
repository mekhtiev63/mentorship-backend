package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/oneonone/application"
)

// AdminHandler serves admin 1:1 endpoints.
type AdminHandler struct {
	admin *application.OneOnOneAdminService
	query *application.OneOnOneQueryService
}

// NewAdminHandler creates AdminHandler.
func NewAdminHandler(admin *application.OneOnOneAdminService, query *application.OneOnOneQueryService) *AdminHandler {
	return &AdminHandler{admin: admin, query: query}
}

// ListRequests GET /admin/one-on-one/requests
func (h *AdminHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	status := r.URL.Query().Get("status")
	var st *string
	if status != "" {
		st = &status
	}
	res, err := h.query.ListAdminRequests(r.Context(), page, perPage, st)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

// GetRequest GET /admin/one-on-one/requests/{requestId}
func (h *AdminHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	dto, err := h.query.GetAdminRequest(r.Context(), chi.URLParam(r, "requestId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

// ApproveRequest POST /admin/one-on-one/requests/{requestId}/approve
func (h *AdminHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	if err := h.admin.ApproveRequest(r.Context(), string(p.UserID), chi.URLParam(r, "requestId")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RejectRequest POST /admin/one-on-one/requests/{requestId}/reject
func (h *AdminHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var req rejectRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	if err := h.admin.RejectRequest(r.Context(), string(p.UserID), chi.URLParam(r, "requestId"), req.Reason); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
