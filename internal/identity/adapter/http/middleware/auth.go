package middleware

import (
	"net/http"
	"strings"

	"github.com/go-mentorship-platform/backend/internal/identity/adapter/httpctx"
	"github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// Authenticate validates Bearer JWT and attaches Principal to context.
func Authenticate(parser domain.TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				httpctx.WriteUnauthorized(w, "missing_token", "authorization required")
				return
			}

			claims, err := parser.Parse(token)
			if err != nil {
				httpctx.WriteUnauthorized(w, "invalid_token", "invalid or expired token")
				return
			}

			principal := domain.Principal{
				UserID:     claims.UserID,
				Roles:      claims.Roles,
				ActiveRole: claims.ActiveRole,
			}
			next.ServeHTTP(w, r.WithContext(httpctx.WithPrincipal(r.Context(), principal)))
		})
	}
}

func bearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	return token, token != ""
}
