package http

import (
	"errors"
	"net/http"

	"github.com/go-mentorship-platform/backend/internal/identity/domain"
	"github.com/go-mentorship-platform/backend/pkg/apperror"
)

func mapError(err error) *apperror.Error {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return apperror.New(apperror.KindUnauthorized, "invalid_credentials", "invalid email or password")
	case errors.Is(err, domain.ErrAccountInactive):
		return apperror.New(apperror.KindUnauthorized, "account_inactive", "account is not active")
	case errors.Is(err, domain.ErrInvalidToken):
		return apperror.New(apperror.KindUnauthorized, "invalid_token", "invalid or expired token")
	case errors.Is(err, domain.ErrActiveRoleNotAllowed):
		return apperror.New(apperror.KindForbidden, "invalid_active_role", "active role is not assigned to the user")
	case errors.Is(err, domain.ErrInvalidRole):
		return apperror.New(apperror.KindValidation, "invalid_role", "unknown role")
	default:
		return apperror.New(apperror.KindInternal, "internal", "internal server error")
	}
}

func writeAppError(w http.ResponseWriter, err error) {
	appErr := mapError(err)
	status := http.StatusInternalServerError
	switch appErr.Kind {
	case apperror.KindUnauthorized:
		status = http.StatusUnauthorized
	case apperror.KindForbidden:
		status = http.StatusForbidden
	case apperror.KindValidation:
		status = http.StatusBadRequest
	case apperror.KindNotFound:
		status = http.StatusNotFound
	}
	writeError(w, status, appErr.Code, appErr.Message)
}
