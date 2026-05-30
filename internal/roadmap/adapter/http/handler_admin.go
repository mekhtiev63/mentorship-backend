package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/roadmap/application"
	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// AdminHandler serves admin roadmap endpoints.
type AdminHandler struct {
	admin *application.AdminService
}

// NewAdminHandler creates AdminHandler.
func NewAdminHandler(admin *application.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

// ListBlocks GET /admin/roadmap/blocks
func (h *AdminHandler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	params := pagination.Normalize(page, perPage)
	filter := domain.AdminBlockFilter{}
	if s := r.URL.Query().Get("status"); s != "" {
		st, err := domain.ParseBlockStatus(s)
		if err != nil {
			writeError(w, err)
			return
		}
		filter.Status = &st
	}
	if a := r.URL.Query().Get("is_active"); a != "" {
		active := a == "true" || a == "1"
		filter.IsActive = &active
	}
	items, total, err := h.admin.ListBlocksAdmin(r.Context(), filter, params)
	if err != nil {
		writeError(w, err)
		return
	}
	meta := pagination.NewMeta(params.Page, params.PageSize, total)
	writeData(w, http.StatusOK, map[string]any{"items": items, "meta": meta})
}

// CreateBlock POST /admin/roadmap/blocks
func (h *AdminHandler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	var req createBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	block, err := h.admin.CreateBlock(r.Context(), application.CreateBlockInput{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, block)
}

// GetBlock GET /admin/roadmap/blocks/{blockId}
func (h *AdminHandler) GetBlock(w http.ResponseWriter, r *http.Request) {
	block, err := h.admin.GetBlockAdmin(r.Context(), chi.URLParam(r, "blockId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, block)
}

// UpdateBlock PATCH /admin/roadmap/blocks/{blockId}
func (h *AdminHandler) UpdateBlock(w http.ResponseWriter, r *http.Request) {
	var req updateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	block, err := h.admin.UpdateBlock(r.Context(), chi.URLParam(r, "blockId"), application.UpdateBlockInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, block)
}

// DeleteBlock DELETE /admin/roadmap/blocks/{blockId}
func (h *AdminHandler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	if err := h.admin.DeleteBlock(r.Context(), chi.URLParam(r, "blockId")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SetBlockActive PATCH /admin/roadmap/blocks/{blockId}/active
func (h *AdminHandler) SetBlockActive(w http.ResponseWriter, r *http.Request) {
	var req setActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	if err := h.admin.SetBlockActive(r.Context(), chi.URLParam(r, "blockId"), req.Active); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PublishBlock POST /admin/roadmap/blocks/{blockId}/publish
func (h *AdminHandler) PublishBlock(w http.ResponseWriter, r *http.Request) {
	block, err := h.admin.PublishBlock(r.Context(), chi.URLParam(r, "blockId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, block)
}

// UnpublishBlock POST /admin/roadmap/blocks/{blockId}/unpublish
func (h *AdminHandler) UnpublishBlock(w http.ResponseWriter, r *http.Request) {
	block, err := h.admin.UnpublishBlock(r.Context(), chi.URLParam(r, "blockId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, block)
}

// ReorderBlocks PUT /admin/roadmap/blocks/reorder
func (h *AdminHandler) ReorderBlocks(w http.ResponseWriter, r *http.Request) {
	var req reorderBlocksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	orders, err := toBlockOrders(req.Items)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.admin.ReorderBlocks(r.Context(), orders); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListMaterials GET /admin/roadmap/blocks/{blockId}/materials
func (h *AdminHandler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	items, err := h.admin.ListMaterialsAdmin(r.Context(), chi.URLParam(r, "blockId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items})
}

// CreateMaterial POST /admin/roadmap/blocks/{blockId}/materials
func (h *AdminHandler) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	var req createMaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	mat, err := h.admin.CreateMaterial(r.Context(), application.CreateMaterialInput{
		BlockID:      chi.URLParam(r, "blockId"),
		Title:        req.Title,
		MaterialType: req.MaterialType,
		URL:          req.URL,
		Required:     req.Required,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, mat)
}

// UpdateMaterial PATCH /admin/roadmap/materials/{materialId}
func (h *AdminHandler) UpdateMaterial(w http.ResponseWriter, r *http.Request) {
	var req updateMaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	mat, err := h.admin.UpdateMaterial(r.Context(), chi.URLParam(r, "materialId"), application.UpdateMaterialInput{
		Title:        req.Title,
		MaterialType: req.MaterialType,
		URL:          req.URL,
		Required:     req.Required,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, mat)
}

// DeleteMaterial DELETE /admin/roadmap/materials/{materialId}
func (h *AdminHandler) DeleteMaterial(w http.ResponseWriter, r *http.Request) {
	if err := h.admin.DeleteMaterial(r.Context(), chi.URLParam(r, "materialId")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SetMaterialActive PATCH /admin/roadmap/materials/{materialId}/active
func (h *AdminHandler) SetMaterialActive(w http.ResponseWriter, r *http.Request) {
	var req setActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	if err := h.admin.SetMaterialActive(r.Context(), chi.URLParam(r, "materialId"), req.Active); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ReorderMaterials PUT /admin/roadmap/blocks/{blockId}/materials/reorder
func (h *AdminHandler) ReorderMaterials(w http.ResponseWriter, r *http.Request) {
	var req reorderMaterialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	orders, err := toMaterialOrders(req.Items)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.admin.ReorderMaterials(r.Context(), chi.URLParam(r, "blockId"), orders); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
