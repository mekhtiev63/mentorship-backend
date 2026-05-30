package domain

import "time"

// RefreshTokenID identifies a stored refresh session.
type RefreshTokenID string

// RefreshToken is a persisted opaque refresh session.
type RefreshToken struct {
	ID        RefreshTokenID
	UserID    UserID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// IsValid reports whether the token can be used.
func (t RefreshToken) IsValid(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	return now.Before(t.ExpiresAt)
}
