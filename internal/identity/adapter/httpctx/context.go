package httpctx

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-mentorship-platform/backend/internal/identity/domain"
)

type principalKey struct{}

// WithPrincipal stores the principal in context.
func WithPrincipal(ctx context.Context, principal domain.Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, principal)
}

// PrincipalFromContext returns the authenticated principal.
func PrincipalFromContext(ctx context.Context) (domain.Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(domain.Principal)
	return p, ok
}

type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// WriteUnauthorized writes 401 JSON.
func WriteUnauthorized(w http.ResponseWriter, code, message string) {
	writeError(w, http.StatusUnauthorized, code, message)
}

// WriteForbidden writes 403 JSON.
func WriteForbidden(w http.ResponseWriter, code, message string) {
	writeError(w, http.StatusForbidden, code, message)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	body := errorEnvelope{}
	body.Error.Code = code
	body.Error.Message = message
	_ = json.NewEncoder(w).Encode(body)
}
