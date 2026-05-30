package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/interview/application"
	"github.com/go-mentorship-platform/backend/internal/interview/domain"
)

// BuddyHandler serves buddy mock interview endpoints.
type BuddyHandler struct {
	mock     *application.MockInterviewService
	feedback *application.InterviewFeedbackService
}

// NewBuddyHandler creates BuddyHandler.
func NewBuddyHandler(mock *application.MockInterviewService, feedback *application.InterviewFeedbackService) *BuddyHandler {
	return &BuddyHandler{mock: mock, feedback: feedback}
}

// CreateMock POST /buddy/interviews/mock
func (h *BuddyHandler) CreateMock(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var body mockCreateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	at, err := parseTime(body.ScheduledAt)
	if err != nil {
		writeBadRequest(w, "validation_error", "invalid scheduled_at")
		return
	}
	dto, err := h.mock.Create(r.Context(), string(p.UserID), application.MockCreateInput{
		StudentID: body.StudentID, ScheduledAt: at, StudentNotes: body.StudentNotes,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, dto)
}

// ListMock GET /buddy/interviews/mock
func (h *BuddyHandler) ListMock(w http.ResponseWriter, r *http.Request) {
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
	res, err := h.mock.List(r.Context(), string(p.UserID), page, perPage, status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

// CompleteMock POST /buddy/interviews/mock/{interviewId}/complete
func (h *BuddyHandler) CompleteMock(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var body feedbackBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	dto, err := h.feedback.CompleteMock(r.Context(), string(p.UserID), chi.URLParam(r, "interviewId"), application.FeedbackInput{
		Feedback: body.Feedback,
		Outcome:  domain.InterviewOutcome(body.Outcome),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}

// GetMockFeedback GET /buddy/interviews/mock/{interviewId}/feedback
func (h *BuddyHandler) GetMockFeedback(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	text, err := h.feedback.GetFeedback(r.Context(), string(p.UserID), chi.URLParam(r, "interviewId"), true)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"feedback": text})
}
