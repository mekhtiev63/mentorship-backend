package security

import (
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/identity/domain"
	"github.com/golang-jwt/jwt/v5"
)

// JWTProvider issues and parses HS256 access tokens.
type JWTProvider struct {
	secret []byte
}

// NewJWTProvider creates a JWTProvider.
func NewJWTProvider(secret string) *JWTProvider {
	return &JWTProvider{secret: []byte(secret)}
}

type accessClaims struct {
	Roles      []string `json:"roles"`
	ActiveRole *string  `json:"active_role,omitempty"`
	jwt.RegisteredClaims
}

// Issue signs an access token with minimal claims.
func (p *JWTProvider) Issue(claims domain.AccessTokenClaims) (string, error) {
	var active *string
	if claims.ActiveRole != nil {
		v := string(*claims.ActiveRole)
		active = &v
	}

	c := accessClaims{
		Roles:      claims.Roles.Strings(),
		ActiveRole: active,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(claims.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ID:        claims.TokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString(p.secret)
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}

// Parse validates and parses an access token.
func (p *JWTProvider) Parse(tokenString string) (domain.AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return p.secret, nil
	})
	if err != nil {
		return domain.AccessTokenClaims{}, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return domain.AccessTokenClaims{}, domain.ErrInvalidToken
	}

	userID, err := domain.ParseUserID(claims.Subject)
	if err != nil {
		return domain.AccessTokenClaims{}, domain.ErrInvalidToken
	}

	roles := make(domain.RoleSet, 0, len(claims.Roles))
	for _, r := range claims.Roles {
		roles = append(roles, domain.Role(r))
	}

	var active *domain.Role
	if claims.ActiveRole != nil {
		role := domain.Role(*claims.ActiveRole)
		active = &role
	}

	var expires time.Time
	if claims.ExpiresAt != nil {
		expires = claims.ExpiresAt.Time
	}

	return domain.AccessTokenClaims{
		UserID:     userID,
		Roles:      roles,
		ActiveRole: active,
		ExpiresAt:  expires,
		TokenID:    claims.ID,
	}, nil
}
