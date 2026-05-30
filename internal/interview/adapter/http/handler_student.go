package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/interview/application"
	"github.com/go-mentorship-platform/backend/internal/interview/domain"
)

// StudentHandler serves student interview endpoints.
type StudentHandler struct {
	real  *application.RealInterviewService
	query *application.InterviewQueryService
	feedback *application.InterviewFeedbackService
}

// NewStudentHandler creates StudentHandler.
func NewStudentHandler(
	real *application.RealInterviewService,
	query *application.InterviewQueryService,
	feedback *application.InterviewFeedbackService,
) *StudentHandler {
	return &StudentHandler{real: real, query: query, feedback: feedback}
}

// CreateReal POST /interviews/real
func (h *StudentHandler) CreateReal(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var body realWriteBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	at, err := parseTime(body.ScheduledAt)
	if err != nil {
		writeBadRequest(w, "validation_error", "invalid scheduled_at")
		return
	}
	dto, err := h.real.Create(r.Context(), string(p.UserID), application.RealCreateInput{
		Company: body.Company, Position: body.Position, ScheduledAt: at,
		StudentNotes: body.StudentNotes, ExternalInterviewer: body.ExternalInterviewer,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, dto)
}

// UpdateReal PATCH /interviews/real/{interviewId}
func (h *StudentHandler) UpdateReal(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var body realWriteBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	at, err := parseTime(body.ScheduledAt)
	if err != nil {
		writeBadRequest(w, "validation_error", "invalid scheduled_at")
		return
	}
	dto, err := h.real.Update(r.Context(), string(p.UserID), chi.URLParam(r, "interviewId"), application.RealUpdateInput{
		Company: body.Company, Position: body.Position, ScheduledAt: at,
		StudentNotes: body.StudentNotes, ExternalInterviewer: body.ExternalInterviewer,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

// CompleteReal POST /interviews/real/{interviewId}/complete
func (h *StudentHandler) CompleteReal(w http.ResponseWriter, r *http.Request) {
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
	dto, err := h.real.Complete(r.Context(), string(p.UserID), chi.URLParam(r, "interviewId"), domain.InterviewOutcome(body.Outcome))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

// ListReal GET /interviews/real
func (h *StudentHandler) ListReal(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	st := r.URL.Query().Get("status")
	var status *string
	if st != "" {
		status = &st
	}
	res, err := h.query.ListRealForStudent(r.Context(), string(p.UserID), page, perPage, status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

// GetReal GET /interviews/real/{interviewId}
func (h *StudentHandler) GetReal(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	dto, err := h.query.GetRealForStudent(r.Context(), string(p.UserID), chi.URLParam(r, "interviewId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

// ListMock GET /interviews/mock
func (h *StudentHandler) ListMock(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	st := r.URL.Query().Get("status")
	var status *string
	if st != "" {
		status = &st
	}
	res, err := h.query.ListMockForStudent(r.Context(), string(p.UserID), page, perPage, status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

// GetMock GET /interviews/mock/{interviewId}
func (h *StudentHandler) GetMock(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	dto, err := h.query.GetMockForStudent(r.Context(), string(p.UserID), chi.URLParam(r, "interviewId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

// GetMockFeedback GET /interviews/mock/{interviewId}/feedback
func (h *StudentHandler) GetMockFeedback(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	text, err := h.feedback.GetFeedback(r.Context(), string(p.UserID), chi.URLParam(r, "interviewId"), false)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"feedback": text})
}

// ListCatalog GET /interviews/catalog
func (h *StudentHandler) ListCatalog(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	company := r.URL.Query().Get("company")
	outcome := r.URL.Query().Get("outcome")
	var c, o *string
	if company != "" {
		c = &company
	}
	if outcome != "" {
		o = &outcome
	}
	res, err := h.query.ListCatalog(r.Context(), page, perPage, c, o)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

func parseTime(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, domain.ErrScheduledRequired
	}
	return time.Parse(time.RFC3339, raw)
}
