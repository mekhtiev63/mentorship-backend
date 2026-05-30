package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Checker verifies dependency readiness.
type Checker interface {
	Ping(ctx context.Context) error
}

// Handler serves liveness and readiness probes.
type Handler struct {
	postgres Checker
	redis    Checker
}

// NewHandler builds a health handler.
func NewHandler(postgres, redis Checker) *Handler {
	return &Handler{
		postgres: postgres,
		redis:    redis,
	}
}

type response struct {
	Status string           `json:"status"`
	Checks map[string]check `json:"checks,omitempty"`
}

type check struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Live returns 200 if the process is running (no dependency checks).
func (h *Handler) Live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, response{Status: "ok"})
}

// Ready returns 200 only when PostgreSQL and Redis are reachable.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	checks := map[string]check{
		"postgres": h.runCheck(ctx, h.postgres),
		"redis":    h.runCheck(ctx, h.redis),
	}

	status := "ok"
	code := http.StatusOK
	for _, c := range checks {
		if c.Status != "ok" {
			status = "degraded"
			code = http.StatusServiceUnavailable
			break
		}
	}

	writeJSON(w, code, response{
		Status: status,
		Checks: checks,
	})
}

func (h *Handler) runCheck(ctx context.Context, c Checker) check {
	if c == nil {
		return check{Status: "ok"}
	}
	if err := c.Ping(ctx); err != nil {
		return check{Status: "fail", Message: err.Error()}
	}
	return check{Status: "ok"}
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
