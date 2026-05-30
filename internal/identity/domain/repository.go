package domain

import "context"

// CredentialsRepository loads authentication data.
type CredentialsRepository interface {
	FindByEmail(ctx context.Context, email Email) (Credentials, error)
	FindByID(ctx context.Context, userID UserID) (Credentials, error)
}

// RoleRepository lists roles assigned to a user.
type RoleRepository interface {
	ListByUser(ctx context.Context, userID UserID) (RoleSet, error)
}

// UserPreferencesRepository stores the active role preference.
type UserPreferencesRepository interface {
	Get(ctx context.Context, userID UserID) (UserPreferences, error)
	SetActiveRole(ctx context.Context, userID UserID, role *Role) error
}

// RefreshTokenRepository persists refresh sessions.
type RefreshTokenRepository interface {
	Store(ctx context.Context, token RefreshToken) error
	FindByHash(ctx context.Context, tokenHash string) (RefreshToken, error)
	Revoke(ctx context.Context, id RefreshTokenID) error
}
