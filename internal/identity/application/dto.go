package application

import "github.com/go-mentorship-platform/backend/internal/identity/domain"

// TokenPair is returned after successful authentication.
type TokenPair struct {
	AccessToken           string
	AccessTokenExpiresIn  int64
	RefreshToken          string
	RequiresRoleSelection bool
}

// UserInfo is the authenticated user snapshot.
type UserInfo struct {
	ID         string
	Email      string
	Status     domain.AccountStatus
	Roles      []string
	ActiveRole *string
}
