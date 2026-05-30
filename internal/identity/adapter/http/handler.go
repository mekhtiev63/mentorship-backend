package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/identity/application"
)

// Handler serves identity HTTP endpoints.
type Handler struct {
	login      *application.LoginService
	logout     *application.LogoutService
	me         *application.MeService
	activeRole *application.ActiveRoleService
}

// NewHandler builds Handler.
func NewHandler(
	login *application.LoginService,
	logout *application.LogoutService,
	me *application.MeService,
	activeRole *application.ActiveRoleService,
) *Handler {
	return &Handler{
		login:      login,
		logout:     logout,
		me:         me,
		activeRole: activeRole,
	}
}

// Login handles POST /auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	pair, user, err := h.login.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeData(w, http.StatusOK, LoginResponse{
		Tokens: toTokenResponse(pair),
		User:   toUserResponse(user),
	})
}

// Logout handles POST /auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	if err := h.logout.Logout(r.Context(), req.RefreshToken); err != nil {
		writeAppError(w, err)
		return
	}

	writeData(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Me handles GET /auth/me.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		httpctx.WriteUnauthorized(w, "missing_principal", "authorization required")
		return
	}

	user, err := h.me.Get(r.Context(), principal.UserID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeData(w, http.StatusOK, MeResponse{User: toUserResponse(user)})
}

// SetActiveRole handles PUT /auth/active-role.
func (h *Handler) SetActiveRole(w http.ResponseWriter, r *http.Request) {
	principal, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		httpctx.WriteUnauthorized(w, "missing_principal", "authorization required")
		return
	}

	var req SetActiveRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	pair, user, err := h.activeRole.SetActiveRole(r.Context(), principal.UserID, req.ActiveRole)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeData(w, http.StatusOK, ActiveRoleResponse{
		Tokens: toTokenResponse(pair),
		User:   toUserResponse(user),
	})
}
