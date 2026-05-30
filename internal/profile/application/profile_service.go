package application

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/profile/domain"
)

// ProfileService manages profile use cases.
type ProfileService struct {
	profiles domain.ProfileRepository
	users    domain.UserExistsReader
	buddies  domain.BuddyAssignmentReader
}

// NewProfileService builds ProfileService.
func NewProfileService(
	profiles domain.ProfileRepository,
	users domain.UserExistsReader,
	buddies domain.BuddyAssignmentReader,
) *ProfileService {
	return &ProfileService{profiles: profiles, users: users, buddies: buddies}
}

// ProfileDTO is API profile shape.
type ProfileDTO struct {
	UserID           string  `json:"user_id"`
	DisplayName      string  `json:"display_name"`
	Bio              string  `json:"bio"`
	AvatarURL        *string `json:"avatar_url"`
	TelegramUsername *string `json:"telegram_username"`
	Visibility       string  `json:"visibility"`
}

// UpdateProfileInput is patch input for own profile.
type UpdateProfileInput struct {
	DisplayName      *string
	Bio              *string
	AvatarURL        *string
	TelegramUsername *string
	Visibility       *string
}

// GetMyProfile returns the owner's profile.
func (s *ProfileService) GetMyProfile(ctx context.Context, userID string) (ProfileDTO, error) {
	id := domain.UserID(userID)
	profile, err := s.profiles.GetByUserID(ctx, id)
	if err != nil {
		return ProfileDTO{}, err
	}
	return toDTO(profile), nil
}

// UpdateMyProfile updates editable fields.
func (s *ProfileService) UpdateMyProfile(ctx context.Context, userID string, in UpdateProfileInput) (ProfileDTO, error) {
	id := domain.UserID(userID)
	profile, err := s.profiles.GetByUserID(ctx, id)
	if err != nil {
		return ProfileDTO{}, err
	}

	if in.DisplayName != nil {
		profile.DisplayName = *in.DisplayName
	}
	if in.Bio != nil {
		profile.Bio = *in.Bio
	}
	if in.AvatarURL != nil {
		profile.AvatarURL = in.AvatarURL
	}
	if in.TelegramUsername != nil {
		tg, err := domain.NormalizeTelegram(*in.TelegramUsername)
		if err != nil {
			return ProfileDTO{}, err
		}
		if tg == "" {
			profile.TelegramUsername = nil
		} else {
			profile.TelegramUsername = &tg
		}
	}
	if in.Visibility != nil {
		v, err := domain.ParseVisibility(*in.Visibility)
		if err != nil {
			return ProfileDTO{}, err
		}
		profile.Visibility = v
	}

	if err := s.profiles.Update(ctx, profile); err != nil {
		return ProfileDTO{}, err
	}
	return toDTO(profile), nil
}

// GetUserProfile returns another user's profile when allowed.
func (s *ProfileService) GetUserProfile(ctx context.Context, viewerID, targetID string, isAdmin bool) (ProfileDTO, error) {
	target := domain.UserID(targetID)
	exists, err := s.users.ExistsActive(ctx, target)
	if err != nil {
		return ProfileDTO{}, err
	}
	if !exists {
		return ProfileDTO{}, domain.ErrNotFound
	}

	profile, err := s.profiles.GetByUserID(ctx, target)
	if err != nil {
		return ProfileDTO{}, err
	}

	rel := &relationshipAdapter{ctx: ctx, reader: s.buddies}
	viewer := domain.ViewerContext{
		ViewerID:     domain.UserID(viewerID),
		IsAdmin:      isAdmin,
		Relationship: rel,
	}
	if !domain.CanView(viewer, target, profile.Visibility) {
		return ProfileDTO{}, domain.ErrForbidden
	}
	return toDTO(profile), nil
}

type relationshipAdapter struct {
	ctx    context.Context
	reader domain.BuddyAssignmentReader
}

func (r *relationshipAdapter) IsAssignedBuddy(buddyID, studentID domain.UserID) bool {
	ok, err := r.reader.IsAssignedBuddy(r.ctx, buddyID, studentID)
	return err == nil && ok
}

func toDTO(p domain.Profile) ProfileDTO {
	return ProfileDTO{
		UserID:           string(p.UserID),
		DisplayName:      p.DisplayName,
		Bio:              p.Bio,
		AvatarURL:        p.AvatarURL,
		TelegramUsername: p.TelegramUsername,
		Visibility:       string(p.Visibility),
	}
}
