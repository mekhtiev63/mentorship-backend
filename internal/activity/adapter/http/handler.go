package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/activity/application"
)

// Handler serves activity HTTP endpoints.
type Handler struct {
	query *application.ActivityQueryService
}

// NewHandler creates Handler.
func NewHandler(query *application.ActivityQueryService) *Handler {
	return &Handler{query: query}
}

// ListMe GET /activity/me
func (h *Handler) ListMe(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	q, err := parseListQuery(r)
	if err != nil {
		writeBadRequest(w, "validation_error", err.Error())
		return
	}
	res, err := h.query.ListMyActivity(r.Context(), string(p.UserID), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

// ListStudent GET /buddy/activity/students/{studentId}
func (h *Handler) ListStudent(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	q, err := parseListQuery(r)
	if err != nil {
		writeBadRequest(w, "validation_error", err.Error())
		return
	}
	res, err := h.query.ListStudentActivity(r.Context(), string(p.UserID), chi.URLParam(r, "studentId"), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

// ListAdmin GET /admin/activity
func (h *Handler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	q, err := parseListQuery(r)
	if err != nil {
		writeBadRequest(w, "validation_error", err.Error())
		return
	}
	if sid := r.URL.Query().Get("subject_user_id"); sid != "" {
		q.SubjectUserID = &sid
	}
	if aid := r.URL.Query().Get("actor_id"); aid != "" {
		q.ActorID = &aid
	}
	res, err := h.query.ListGlobal(r.Context(), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

func parseListQuery(r *http.Request) (application.ListQuery, error) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	q := application.ListQuery{Page: page, PageSize: perPage}
	if t := r.URL.Query().Get("activity_type"); t != "" {
		q.ActivityType = &t
	}
	if v := r.URL.Query().Get("verb"); v != "" {
		q.Verb = &v
	}
	if o := r.URL.Query().Get("object_type"); o != "" {
		q.ObjectType = &o
	}
	if raw := r.URL.Query().Get("from"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return q, err
		}
		q.From = &t
	}
	if raw := r.URL.Query().Get("to"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return q, err
		}
		q.To = &t
	}
	return q, nil
}
