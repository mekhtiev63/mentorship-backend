package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/calendar/application"
	"github.com/go-mentorship-platform/backend/internal/calendar/domain"
)

// Handler serves calendar HTTP endpoints.
type Handler struct {
	events *application.CalendarEventService
	query  *application.CalendarQueryService
}

// NewHandler creates Handler.
func NewHandler(events *application.CalendarEventService, query *application.CalendarQueryService) *Handler {
	return &Handler{events: events, query: query}
}

// ListEvents GET /calendar/events
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
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
	res, err := h.query.List(r.Context(), string(p.UserID), q, isAdmin(p))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

// ListUpcoming GET /calendar/events/upcoming
func (h *Handler) ListUpcoming(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	res, err := h.query.ListUpcoming(r.Context(), string(p.UserID), page, perPage)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

// CreateEvent POST /calendar/events
func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var body eventWriteBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	start, end, rel, err := parseWriteBody(body)
	if err != nil {
		writeBadRequest(w, "validation_error", err.Error())
		return
	}
	dto, err := h.events.Create(r.Context(), string(p.UserID), application.CreateEventInput{
		Title: body.Title, Description: body.Description, StartsAt: start, EndsAt: end,
		RelatedType: rel, RelatedID: body.RelatedID, AttendeeIDs: body.AttendeeIDs,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, dto)
}

// GetEvent GET /calendar/events/{eventId}
func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	dto, err := h.query.Get(r.Context(), string(p.UserID), chi.URLParam(r, "eventId"), isAdmin(p))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

// UpdateEvent PATCH /calendar/events/{eventId}
func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var body eventUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	start, err := time.Parse(time.RFC3339, body.StartsAt)
	if err != nil {
		writeBadRequest(w, "validation_error", "invalid starts_at")
		return
	}
	end, err := time.Parse(time.RFC3339, body.EndsAt)
	if err != nil {
		writeBadRequest(w, "validation_error", "invalid ends_at")
		return
	}
	dto, err := h.events.Update(r.Context(), string(p.UserID), chi.URLParam(r, "eventId"), application.UpdateEventInput{
		Title: body.Title, Description: body.Description, StartsAt: start, EndsAt: end, AttendeeIDs: body.AttendeeIDs,
	}, isAdmin(p))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

// CancelEvent POST /calendar/events/{eventId}/cancel
func (h *Handler) CancelEvent(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	dto, err := h.events.Cancel(r.Context(), string(p.UserID), chi.URLParam(r, "eventId"), isAdmin(p))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

// DeleteEvent DELETE /calendar/events/{eventId}
func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	if err := h.events.Delete(r.Context(), string(p.UserID), chi.URLParam(r, "eventId"), isAdmin(p)); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func isAdmin(p identitydomain.Principal) bool {
	return p.HasRole(identitydomain.RoleAdmin)
}

func parseListQuery(r *http.Request) (application.ListEventsQuery, error) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	q := application.ListEventsQuery{
		Page: page, PageSize: perPage,
		IncludeCancelled: r.URL.Query().Get("include_cancelled") == "true",
		IncludeDeleted:   r.URL.Query().Get("include_deleted") == "true",
	}
	if rt := r.URL.Query().Get("related_type"); rt != "" {
		q.RelatedType = &rt
	}
	if raw := r.URL.Query().Get("from"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return q, domain.ErrInvalidTimeRange
		}
		q.From = &t
	}
	if raw := r.URL.Query().Get("to"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return q, domain.ErrInvalidTimeRange
		}
		q.To = &t
	}
	return q, nil
}

func parseWriteBody(body eventWriteBody) (time.Time, time.Time, domain.RelatedType, error) {
	start, err := time.Parse(time.RFC3339, body.StartsAt)
	if err != nil {
		return time.Time{}, time.Time{}, "", domain.ErrInvalidTimeRange
	}
	end, err := time.Parse(time.RFC3339, body.EndsAt)
	if err != nil {
		return time.Time{}, time.Time{}, "", domain.ErrInvalidTimeRange
	}
	rel, _ := domain.ParseRelatedType(body.RelatedType)
	if body.RelatedType == "" {
		rel = domain.RelatedOther
	}
	return start, end, rel, nil
}
