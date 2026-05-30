package domain

import "time"

// AccessTokenClaims are minimal JWT claims issued by the platform.
type AccessTokenClaims struct {
	UserID     UserID
	Roles      RoleSet
	ActiveRole *Role
	ExpiresAt  time.Time
	TokenID    string
}

// TokenIssuer signs access tokens.
type TokenIssuer interface {
	Issue(claims AccessTokenClaims) (string, error)
}

// TokenParser validates access tokens.
type TokenParser interface {
	Parse(token string) (AccessTokenClaims, error)
}

// RefreshTokenGenerator creates opaque refresh tokens.
type RefreshTokenGenerator interface {
	Generate() (raw string, hash string, err error)
}
