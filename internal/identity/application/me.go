package application

import (
	"context"
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// MeService returns the current user profile for authentication context.
type MeService struct {
	credentials domain.CredentialsRepository
	roles       domain.RoleRepository
	preferences domain.UserPreferencesRepository
}

// NewMeService builds MeService.
func NewMeService(
	credentials domain.CredentialsRepository,
	roles domain.RoleRepository,
	preferences domain.UserPreferencesRepository,
) *MeService {
	return &MeService{
		credentials: credentials,
		roles:       roles,
		preferences: preferences,
	}
}

// Get returns user info for the principal.
func (s *MeService) Get(ctx context.Context, userID domain.UserID) (UserInfo, error) {
	creds, err := s.credentials.FindByID(ctx, userID)
	if err != nil {
		return UserInfo{}, err
	}
	if !creds.Account.CanLogin() {
		return UserInfo{}, domain.ErrAccountInactive
	}

	roleSet, err := s.roles.ListByUser(ctx, userID)
	if err != nil {
		return UserInfo{}, fmt.Errorf("list roles: %w", err)
	}

	prefs, err := s.preferences.Get(ctx, userID)
	if err != nil {
		return UserInfo{}, fmt.Errorf("get preferences: %w", err)
	}

	var active *domain.Role
	if prefs.ActiveRole != nil && roleSet.Contains(*prefs.ActiveRole) {
		active = prefs.ActiveRole
	}

	return toUserInfo(creds, roleSet, active), nil
}
