package middleware

import (
	"net/http"

	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// RequireActiveRole ensures the JWT active_role claim matches allowed roles.
func RequireActiveRole(allowed ...domain.Role) func(http.Handler) http.Handler {
	allowedSet := make(map[domain.Role]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := httpctx.PrincipalFromContext(r.Context())
			if !ok {
				httpctx.WriteUnauthorized(w, "missing_principal", "authorization required")
				return
			}
			if principal.ActiveRole == nil {
				httpctx.WriteForbidden(w, "active_role_required", "active role must be selected")
				return
			}
			if _, ok := allowedSet[*principal.ActiveRole]; !ok {
				httpctx.WriteForbidden(w, "forbidden", "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole ensures the principal has one of the roles in JWT roles claim.
func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := httpctx.PrincipalFromContext(r.Context())
			if !ok {
				httpctx.WriteUnauthorized(w, "missing_principal", "authorization required")
				return
			}
			for _, required := range roles {
				if principal.HasRole(required) {
					next.ServeHTTP(w, r)
					return
				}
			}
			httpctx.WriteForbidden(w, "forbidden", "insufficient permissions")
		})
	}
}
