package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-mentorship-platform/backend/internal/interview/domain"
	"github.com/go-mentorship-platform/backend/pkg/apperror"
)

type envelope struct {
	Data any `json:"data"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, envelope{Data: data})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeBadRequest(w http.ResponseWriter, code, message string) {
	writeJSON(w, http.StatusBadRequest, errorEnvelope{Error: errorBody{Code: code, Message: message}})
}

func writeError(w http.ResponseWriter, err error) {
	appErr := mapDomain(err)
	status := http.StatusInternalServerError
	switch appErr.Kind {
	case apperror.KindNotFound:
		status = http.StatusNotFound
	case apperror.KindConflict:
		status = http.StatusConflict
	case apperror.KindValidation:
		status = http.StatusBadRequest
	case apperror.KindForbidden:
		status = http.StatusForbidden
	}
	writeJSON(w, status, errorEnvelope{Error: errorBody{Code: appErr.Code, Message: appErr.Message}})
}

func mapDomain(err error) *apperror.Error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return apperror.New(apperror.KindNotFound, "not_found", "interview not found")
	case errors.Is(err, domain.ErrForbidden):
		return apperror.New(apperror.KindForbidden, "forbidden", "forbidden")
	case errors.Is(err, domain.ErrInvalidTransition):
		return apperror.New(apperror.KindConflict, "invalid_transition", err.Error())
	case errors.Is(err, domain.ErrFeedbackRequired), errors.Is(err, domain.ErrCompanyRequired),
		errors.Is(err, domain.ErrPositionRequired), errors.Is(err, domain.ErrScheduledRequired),
		errors.Is(err, domain.ErrInvalidOutcome):
		return apperror.New(apperror.KindValidation, "validation_error", err.Error())
	default:
		return apperror.New(apperror.KindInternal, "internal", "internal server error")
	}
}
