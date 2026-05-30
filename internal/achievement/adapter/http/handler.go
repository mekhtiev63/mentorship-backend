package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	identitydomain "github.com/go-mentorship-platform/backend/internal/identity/domain"
	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/achievement/application"
)

// Handler serves achievement HTTP endpoints.
type Handler struct {
	catalog *application.AchievementCatalogService
}

// NewHandler creates Handler.
func NewHandler(catalog *application.AchievementCatalogService) *Handler {
	return &Handler{catalog: catalog}
}

// ListCatalog GET /achievements
func (h *Handler) ListCatalog(w http.ResponseWriter, r *http.Request) {
	items, err := h.catalog.ListCatalog(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items})
}

// ListMyAchievements GET /me/achievements
func (h *Handler) ListMyAchievements(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	items, err := h.catalog.ListUserAchievements(r.Context(), string(p.UserID), string(p.UserID), p.HasRole(identitydomain.RoleAdmin))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items})
}

// ListUserAchievements GET /users/{userId}/achievements
func (h *Handler) ListUserAchievements(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	items, err := h.catalog.ListUserAchievements(r.Context(), string(p.UserID), chi.URLParam(r, "userId"), p.HasRole(identitydomain.RoleAdmin))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items})
}
