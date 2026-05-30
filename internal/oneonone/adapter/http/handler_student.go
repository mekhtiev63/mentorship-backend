package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/oneonone/application"
)

// StudentHandler serves student 1:1 endpoints.
type StudentHandler struct {
	cmd   *application.OneOnOneService
	query *application.OneOnOneQueryService
}

// NewStudentHandler creates StudentHandler.
func NewStudentHandler(cmd *application.OneOnOneService, query *application.OneOnOneQueryService) *StudentHandler {
	return &StudentHandler{cmd: cmd, query: query}
}

// CreateRequest POST /one-on-one/requests
func (h *StudentHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var req createRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	dto, err := h.cmd.CreateRequest(r.Context(), string(p.UserID), application.CreateRequestInput{
		Message:        req.Message,
		PreferredSlots: req.PreferredSlots,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, dto)
}

// ListRequests GET /one-on-one/requests
func (h *StudentHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	res, err := h.query.ListStudentRequests(r.Context(), string(p.UserID), page, perPage)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

// GetRequest GET /one-on-one/requests/{requestId}
func (h *StudentHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	dto, err := h.query.GetStudentRequest(r.Context(), string(p.UserID), chi.URLParam(r, "requestId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

// CancelRequest POST /one-on-one/requests/{requestId}/cancel
func (h *StudentHandler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	if err := h.cmd.CancelRequest(r.Context(), string(p.UserID), chi.URLParam(r, "requestId")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
