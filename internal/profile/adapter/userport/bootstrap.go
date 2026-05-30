package userport

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/profile/domain"
	userdomain "github.com/go-mentorship-platform/backend/internal/user/domain"
)

// ProfileBootstrap adapts profile repository for the user module.
type ProfileBootstrap struct {
	profiles domain.ProfileRepository
}

// NewProfileBootstrap creates ProfileBootstrap.
func NewProfileBootstrap(profiles domain.ProfileRepository) *ProfileBootstrap {
	return &ProfileBootstrap{profiles: profiles}
}

// EnsureEmpty implements userdomain.ProfileBootstrap.
func (b *ProfileBootstrap) EnsureEmpty(ctx context.Context, userID userdomain.UserID) error {
	return b.profiles.EnsureEmpty(ctx, domain.UserID(userID))
}
