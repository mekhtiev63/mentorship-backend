package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	identityhttpctx "github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
	"github.com/go-mentorship-platform/backend/internal/profile/application"
	"github.com/go-mentorship-platform/backend/internal/profile/domain"
	"github.com/go-mentorship-platform/backend/pkg/apperror"
)

// Handler serves profile HTTP endpoints.
type Handler struct {
	profiles *application.ProfileService
}

// NewHandler creates Handler.
func NewHandler(profiles *application.ProfileService) *Handler {
	return &Handler{profiles: profiles}
}

// GetMyProfile GET /me/profile
func (h *Handler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	principal, ok := identityhttpctx.PrincipalFromContext(r.Context())
	if !ok {
		identityhttpctx.WriteUnauthorized(w, "unauthorized", "authorization required")
		return
	}
	profile, err := h.profiles.GetMyProfile(r.Context(), principal.UserID.String())
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"profile": profile})
}

// PatchMyProfile PATCH /me/profile
func (h *Handler) PatchMyProfile(w http.ResponseWriter, r *http.Request) {
	principal, ok := identityhttpctx.PrincipalFromContext(r.Context())
	if !ok {
		identityhttpctx.WriteUnauthorized(w, "unauthorized", "authorization required")
		return
	}
	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	profile, err := h.profiles.UpdateMyProfile(r.Context(), principal.UserID.String(), application.UpdateProfileInput{
		DisplayName:      req.DisplayName,
		Bio:              req.Bio,
		AvatarURL:        req.AvatarURL,
		TelegramUsername: req.TelegramUsername,
		Visibility:       req.Visibility,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"profile": profile})
}

// GetUserProfile GET /users/{userId}/profile
func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	principal, ok := identityhttpctx.PrincipalFromContext(r.Context())
	if !ok {
		identityhttpctx.WriteUnauthorized(w, "unauthorized", "authorization required")
		return
	}
	isAdmin := principal.ActiveRole != nil && *principal.ActiveRole == identitydomain.RoleAdmin
	profile, err := h.profiles.GetUserProfile(r.Context(), principal.UserID.String(), chi.URLParam(r, "userId"), isAdmin)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"profile": profile})
}

type updateProfileRequest struct {
	DisplayName      *string `json:"display_name"`
	Bio              *string `json:"bio"`
	AvatarURL        *string `json:"avatar_url"`
	TelegramUsername *string `json:"telegram_username"`
	Visibility       *string `json:"visibility"`
}

type envelope struct {
	Data any `json:"data"`
}

type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: data})
}

func writeBadRequest(w http.ResponseWriter, code, msg string) {
	writeApp(w, http.StatusBadRequest, apperror.New(apperror.KindValidation, code, msg))
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
	writeApp(w, status, appErr)
}

func writeApp(w http.ResponseWriter, status int, appErr *apperror.Error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	body := errorEnvelope{}
	body.Error.Code = appErr.Code
	body.Error.Message = appErr.Message
	_ = json.NewEncoder(w).Encode(body)
}

func mapDomain(err error) *apperror.Error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return apperror.New(apperror.KindNotFound, "not_found", "profile not found")
	case errors.Is(err, domain.ErrTelegramTaken):
		return apperror.New(apperror.KindConflict, "telegram_taken", "telegram username already taken")
	case errors.Is(err, domain.ErrInvalidVisibility), errors.Is(err, domain.ErrInvalidTelegram):
		return apperror.New(apperror.KindValidation, "validation_error", err.Error())
	case errors.Is(err, domain.ErrForbidden):
		return apperror.New(apperror.KindForbidden, "forbidden", "forbidden")
	default:
		return apperror.New(apperror.KindInternal, "internal", "internal server error")
	}
}
