package application

import (
	"context"
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/identity/domain"
)

// LogoutService revokes refresh sessions.
type LogoutService struct {
	refreshStore domain.RefreshTokenRepository
}

// NewLogoutService builds LogoutService.
func NewLogoutService(refreshStore domain.RefreshTokenRepository) *LogoutService {
	return &LogoutService{refreshStore: refreshStore}
}

// Logout revokes the refresh token identified by its hash.
func (s *LogoutService) Logout(ctx context.Context, refreshTokenRaw string) error {
	if refreshTokenRaw == "" {
		return domain.ErrInvalidToken
	}

	_, hash, err := domain.HashRefreshToken(refreshTokenRaw)
	if err != nil {
		return domain.ErrInvalidToken
	}

	stored, err := s.refreshStore.FindByHash(ctx, hash)
	if err != nil {
		return domain.ErrInvalidToken
	}

	if err := s.refreshStore.Revoke(ctx, stored.ID); err != nil {
		return fmt.Errorf("revoke refresh: %w", err)
	}
	return nil
}
