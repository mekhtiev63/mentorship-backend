package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/notification/application"
)

// Handler serves notification HTTP endpoints.
type Handler struct {
	query *application.NotificationQueryService
}

// NewHandler creates Handler.
func NewHandler(query *application.NotificationQueryService) *Handler {
	return &Handler{query: query}
}

// List GET /notifications
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
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
	res, err := h.query.ListMine(r.Context(), string(p.UserID), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

// UnreadCount GET /notifications/unread-count
func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	res, err := h.query.GetUnreadCount(r.Context(), string(p.UserID))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

// MarkRead PATCH /notifications/{notificationId}/read
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	if err := h.query.MarkRead(r.Context(), string(p.UserID), chi.URLParam(r, "notificationId")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MarkAllRead POST /notifications/read-all
func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	res, err := h.query.MarkAllRead(r.Context(), string(p.UserID))
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
	if rs := r.URL.Query().Get("read_status"); rs != "" {
		q.ReadStatus = &rs
	}
	if t := r.URL.Query().Get("notification_type"); t != "" {
		q.NotificationType = &t
	}
	return q, nil
}
