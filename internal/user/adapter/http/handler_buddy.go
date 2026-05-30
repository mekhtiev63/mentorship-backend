package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	identityhttpctx "github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/user/application"
)

// BuddyHandler serves buddy endpoints.
type BuddyHandler struct {
	buddy *application.BuddyService
}

// NewBuddyHandler creates BuddyHandler.
func NewBuddyHandler(buddy *application.BuddyService) *BuddyHandler {
	return &BuddyHandler{buddy: buddy}
}

// ListStudents GET /buddy/students
func (h *BuddyHandler) ListStudents(w http.ResponseWriter, r *http.Request) {
	principal, ok := identityhttpctx.PrincipalFromContext(r.Context())
	if !ok {
		identityhttpctx.WriteUnauthorized(w, "unauthorized", "authorization required")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	students, meta, err := h.buddy.ListStudents(r.Context(), principal.UserID.String(), page, perPage)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": students, "meta": meta})
}

// GetStudent GET /buddy/students/{studentId}
func (h *BuddyHandler) GetStudent(w http.ResponseWriter, r *http.Request) {
	principal, ok := identityhttpctx.PrincipalFromContext(r.Context())
	if !ok {
		identityhttpctx.WriteUnauthorized(w, "unauthorized", "authorization required")
		return
	}
	studentID := chi.URLParam(r, "studentId")
	if err := h.buddy.EnsureAssigned(r.Context(), principal.UserID.String(), studentID); err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"id": studentID})
}
