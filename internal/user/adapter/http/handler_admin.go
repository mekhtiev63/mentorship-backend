package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/user/application"
)

// AdminHandler serves admin user endpoints.
type AdminHandler struct {
	admin *application.AdminService
}

// NewAdminHandler creates AdminHandler.
func NewAdminHandler(admin *application.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

// ListUsers GET /admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	users, meta, err := h.admin.ListUsers(r.Context(), page, perPage, r.URL.Query().Get("email"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": users, "meta": meta})
}

// CreateUser POST /admin/users
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	user, err := h.admin.CreateUser(r.Context(), application.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
		Roles:    req.Roles,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, user)
}

// GetUser GET /admin/users/{userId}
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.admin.GetUser(r.Context(), chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, user)
}

// UpdateStatus PATCH /admin/users/{userId}/status
func (h *AdminHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	user, err := h.admin.UpdateStatus(r.Context(), chi.URLParam(r, "userId"), req.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, user)
}

// ReplaceRoles PUT /admin/users/{userId}/roles
func (h *AdminHandler) ReplaceRoles(w http.ResponseWriter, r *http.Request) {
	var req replaceRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	user, err := h.admin.ReplaceRoles(r.Context(), chi.URLParam(r, "userId"), req.Roles)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, user)
}

// DeleteUser DELETE /admin/users/{userId}
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if err := h.admin.SoftDeleteUser(r.Context(), chi.URLParam(r, "userId")); err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CreateBuddyAssignment POST /admin/buddy-assignments
func (h *AdminHandler) CreateBuddyAssignment(w http.ResponseWriter, r *http.Request) {
	var req buddyAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	if err := h.admin.CreateBuddyAssignment(r.Context(), req.StudentID, req.BuddyID); err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// DeleteBuddyAssignment DELETE /admin/buddy-assignments/{assignmentId}
func (h *AdminHandler) DeleteBuddyAssignment(w http.ResponseWriter, r *http.Request) {
	if err := h.admin.DeleteBuddyAssignment(r.Context(), chi.URLParam(r, "assignmentId")); err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"status": "ok"})
}
