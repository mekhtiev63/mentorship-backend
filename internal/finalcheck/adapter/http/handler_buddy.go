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

// BuddyHandler serves buddy final-check endpoints.
type BuddyHandler struct {
	svc   *application.FinalCheckService
	query *application.FinalCheckQueryService
}

// NewBuddyHandler creates BuddyHandler.
func NewBuddyHandler(svc *application.FinalCheckService, query *application.FinalCheckQueryService) *BuddyHandler {
	return &BuddyHandler{svc: svc, query: query}
}

// GetStudent GET /buddy/final-check/students/{studentId}
func (h *BuddyHandler) GetStudent(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	dto, err := h.query.GetStudentStatus(r.Context(), string(p.UserID), chi.URLParam(r, "studentId"), false)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

func (h *BuddyHandler) schedule(w http.ResponseWriter, r *http.Request, kind domain.CheckKind) {
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
	dto, err := h.svc.Schedule(r.Context(), string(p.UserID), chi.URLParam(r, "studentId"), kind, at, false)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

func (h *BuddyHandler) complete(w http.ResponseWriter, r *http.Request, kind domain.CheckKind) {
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
	dto, err := h.svc.Complete(r.Context(), string(p.UserID), chi.URLParam(r, "studentId"), kind, body.Feedback, false)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

func (h *BuddyHandler) fail(w http.ResponseWriter, r *http.Request, kind domain.CheckKind) {
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
	dto, err := h.svc.Fail(r.Context(), string(p.UserID), chi.URLParam(r, "studentId"), kind, body.Reason, false)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

// ScheduleTech POST .../tech/schedule
func (h *BuddyHandler) ScheduleTech(w http.ResponseWriter, r *http.Request) {
	h.schedule(w, r, domain.CheckTech)
}

// CompleteTech POST .../tech/complete
func (h *BuddyHandler) CompleteTech(w http.ResponseWriter, r *http.Request) {
	h.complete(w, r, domain.CheckTech)
}

// FailTech POST .../tech/fail
func (h *BuddyHandler) FailTech(w http.ResponseWriter, r *http.Request) {
	h.fail(w, r, domain.CheckTech)
}

// ScheduleRoast POST .../roast/schedule
func (h *BuddyHandler) ScheduleRoast(w http.ResponseWriter, r *http.Request) {
	h.schedule(w, r, domain.CheckRoast)
}

// CompleteRoast POST .../roast/complete
func (h *BuddyHandler) CompleteRoast(w http.ResponseWriter, r *http.Request) {
	h.complete(w, r, domain.CheckRoast)
}

// FailRoast POST .../roast/fail
func (h *BuddyHandler) FailRoast(w http.ResponseWriter, r *http.Request) {
	h.fail(w, r, domain.CheckRoast)
}
