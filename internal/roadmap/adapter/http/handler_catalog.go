package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-mentorship-platform/backend/internal/roadmap/application"
)

// CatalogHandler serves public roadmap reads.
type CatalogHandler struct {
	catalog *application.CatalogService
}

// NewCatalogHandler creates CatalogHandler.
func NewCatalogHandler(catalog *application.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalog: catalog}
}

// ListBlocks GET /roadmap/blocks
func (h *CatalogHandler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	blocks, err := h.catalog.ListPublishedBlocks(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": blocks})
}

// GetBlock GET /roadmap/blocks/{blockId}
func (h *CatalogHandler) GetBlock(w http.ResponseWriter, r *http.Request) {
	block, err := h.catalog.GetPublishedBlock(r.Context(), chi.URLParam(r, "blockId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, block)
}
