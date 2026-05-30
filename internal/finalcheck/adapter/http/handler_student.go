package http

import (
	"net/http"

	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/finalcheck/application"
)

// StudentHandler serves student final-check endpoints.
type StudentHandler struct {
	query *application.FinalCheckQueryService
}

// NewStudentHandler creates StudentHandler.
func NewStudentHandler(query *application.FinalCheckQueryService) *StudentHandler {
	return &StudentHandler{query: query}
}

// GetMe GET /final-check/me
func (h *StudentHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	dto, err := h.query.GetMyStatus(r.Context(), string(p.UserID))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, dto)
}
