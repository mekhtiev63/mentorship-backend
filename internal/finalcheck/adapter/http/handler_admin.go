package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/finalcheck/application"
	"github.com/go-mentorship-platform/backend/internal/finalcheck/domain"
)

// AdminHandler serves admin final-check endpoints.
type AdminHandler struct {
	svc   *application.FinalCheckService
	query *application.FinalCheckQueryService
}

// NewAdminHandler creates AdminHandler.
func NewAdminHandler(svc *application.FinalCheckService, query *application.FinalCheckQueryService) *AdminHandler {
	return &AdminHandler{svc: svc, query: query}
}

// GetStudent GET /admin/final-check/students/{studentId}
func (h *AdminHandler) GetStudent(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	dto, err := h.query.GetStudentStatus(r.Context(), string(p.UserID), chi.URLParam(r, "studentId"), true)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

func (h *AdminHandler) schedule(w http.ResponseWriter, r *http.Request, kind domain.CheckKind) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var body scheduleBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	at, err := time.Parse(time.RFC3339, body.ScheduledAt)
	if err != nil {
		writeBadRequest(w, "validation_error", "invalid scheduled_at")
		return
	}
	dto, err := h.svc.Schedule(r.Context(), string(p.UserID), chi.URLParam(r, "studentId"), kind, at, true)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

func (h *AdminHandler) complete(w http.ResponseWriter, r *http.Request, kind domain.CheckKind) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var body completeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	dto, err := h.svc.Complete(r.Context(), string(p.UserID), chi.URLParam(r, "studentId"), kind, body.Feedback, true)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

func (h *AdminHandler) fail(w http.ResponseWriter, r *http.Request, kind domain.CheckKind) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var body failBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	dto, err := h.svc.Fail(r.Context(), string(p.UserID), chi.URLParam(r, "studentId"), kind, body.Reason, true)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

func (h *AdminHandler) ScheduleTech(w http.ResponseWriter, r *http.Request) {
	h.schedule(w, r, domain.CheckTech)
}
func (h *AdminHandler) CompleteTech(w http.ResponseWriter, r *http.Request) {
	h.complete(w, r, domain.CheckTech)
}
func (h *AdminHandler) FailTech(w http.ResponseWriter, r *http.Request) {
	h.fail(w, r, domain.CheckTech)
}
func (h *AdminHandler) ScheduleRoast(w http.ResponseWriter, r *http.Request) {
	h.schedule(w, r, domain.CheckRoast)
}
func (h *AdminHandler) CompleteRoast(w http.ResponseWriter, r *http.Request) {
	h.complete(w, r, domain.CheckRoast)
}
func (h *AdminHandler) FailRoast(w http.ResponseWriter, r *http.Request) {
	h.fail(w, r, domain.CheckRoast)
}
