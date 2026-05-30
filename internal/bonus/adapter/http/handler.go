package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/bonus/application"
)

// Handler serves bonus HTTP endpoints.
type Handler struct {
	balance *application.BonusBalanceQueryService
	convert *application.DiscountConversionService
}

// NewHandler creates Handler.
func NewHandler(
	balance *application.BonusBalanceQueryService,
	convert *application.DiscountConversionService,
) *Handler {
	return &Handler{balance: balance, convert: convert}
}

// GetBonus GET /me/bonus
func (h *Handler) GetBonus(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	data, err := h.balance.GetBalance(r.Context(), string(p.UserID))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, data)
}

// ListTransactions GET /me/bonus/transactions
func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	items, meta, err := h.balance.ListTransactions(r.Context(), string(p.UserID), page, perPage)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items, "meta": meta})
}

// Convert POST /me/bonus/convert
func (h *Handler) Convert(w http.ResponseWriter, r *http.Request) {
	p, ok := httpctx.PrincipalFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "unauthorized", "authorization required")
		return
	}
	var req convertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid_json", "invalid body")
		return
	}
	key := r.Header.Get("Idempotency-Key")
	res, err := h.convert.Convert(r.Context(), string(p.UserID), req.Points, key)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}
