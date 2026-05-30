package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-mentorship-platform/backend/internal/bonus/domain"
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
	case apperror.KindValidation:
		status = http.StatusBadRequest
	case apperror.KindConflict:
		status = http.StatusConflict
	}
	writeJSON(w, status, errorEnvelope{Error: errorBody{Code: appErr.Code, Message: appErr.Message}})
}

func mapDomain(err error) *apperror.Error {
	switch {
	case errors.Is(err, domain.ErrInsufficientBalance):
		return apperror.New(apperror.KindValidation, "insufficient_balance", err.Error())
	case errors.Is(err, domain.ErrDiscountLimit):
		return apperror.New(apperror.KindValidation, "discount_limit_exceeded", err.Error())
	case errors.Is(err, domain.ErrInvalidAmount), errors.Is(err, domain.ErrInvalidIdempotencyKey):
		return apperror.New(apperror.KindValidation, "validation_error", err.Error())
	case errors.Is(err, domain.ErrDuplicateOperation):
		return apperror.New(apperror.KindConflict, "duplicate", err.Error())
	default:
		return apperror.New(apperror.KindInternal, "internal", "internal server error")
	}
}
