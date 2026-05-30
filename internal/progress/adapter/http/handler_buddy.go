package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/progress/application"
)

// BuddyHandler serves buddy progress endpoints.
type BuddyHandler struct {
	approval *application.BlockApprovalService
	query    *application.BuddyStudentsQueryService
}

// NewBuddyHandler creates BuddyHandler.
func NewBuddyHandler(
	approval *application.BlockApprovalService,
	query *application.BuddyStudentsQueryService,
) *BuddyHandler {
	return &BuddyHandler{approval: approval, query: query}
}

func buddyIDFromRequest(r *http.Request) (string, bool) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		return "", false
	}
	return string(p.UserID), true
}

// ListStudents GET /buddy/progress/students
func (h *BuddyHandler) ListStudents(w http.ResponseWriter, r *http.Request) {
	bid, ok := buddyIDFromRequest(r)
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	items, meta, err := h.query.ListStudentsWithProgress(r.Context(), bid, page, perPage)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items, "meta": meta})
}

// GetStudentProgress GET /buddy/progress/students/{studentId}
func (h *BuddyHandler) GetStudentProgress(w http.ResponseWriter, r *http.Request) {
	bid, ok := buddyIDFromRequest(r)
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	items, err := h.query.GetStudentProgress(r.Context(), bid, chi.URLParam(r, "studentId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items})
}

// ListApprovals GET /buddy/progress/approvals
func (h *BuddyHandler) ListApprovals(w http.ResponseWriter, r *http.Request) {
	bid, ok := buddyIDFromRequest(r)
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	items, meta, err := h.query.ListApprovalQueue(r.Context(), bid, page, perPage)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items, "meta": meta})
}

// ApproveBlock POST /buddy/progress/students/{studentId}/blocks/{blockId}/approve
func (h *BuddyHandler) ApproveBlock(w http.ResponseWriter, r *http.Request) {
	bid, ok := buddyIDFromRequest(r)
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	if err := h.approval.ApproveBlockAsBuddy(r.Context(), bid, chi.URLParam(r, "studentId"), chi.URLParam(r, "blockId")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RejectBlock POST /buddy/progress/students/{studentId}/blocks/{blockId}/reject
func (h *BuddyHandler) RejectBlock(w http.ResponseWriter, r *http.Request) {
	bid, ok := buddyIDFromRequest(r)
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var req rejectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	if err := h.approval.RejectBlockAsBuddy(r.Context(), bid, chi.URLParam(r, "studentId"), chi.URLParam(r, "blockId"), req.Reason); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
