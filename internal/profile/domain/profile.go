package domain

import (
	"regexp"
	"strings"
	"time"
)

// UserID references the profile owner.
type UserID string

// ParseUserID parses UUID user id.
func ParseUserID(raw string) (UserID, error) {
	if raw == "" {
		return "", ErrNotFound
	}
	return UserID(raw), nil
}

// Visibility controls profile read access.
type Visibility string

const (
	VisibilityPublic      Visibility = "public"
	VisibilityBuddiesOnly Visibility = "buddies_only"
	VisibilityPrivate     Visibility = "private"
)

// ParseVisibility parses visibility enum.
func ParseVisibility(raw string) (Visibility, error) {
	switch Visibility(raw) {
	case VisibilityPublic, VisibilityBuddiesOnly, VisibilityPrivate:
		return Visibility(raw), nil
	default:
		return "", ErrInvalidVisibility
	}
}

var telegramRe = regexp.MustCompile(`^[a-z0-9_]{5,32}$`)

// Profile is the profile aggregate root.
type Profile struct {
	UserID           UserID
	DisplayName      string
	Bio              string
	AvatarURL        *string
	TelegramUsername *string
	Visibility       Visibility
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NormalizeTelegram validates and normalizes telegram username (without @).
func NormalizeTelegram(raw string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(raw, "@")))
	if v == "" {
		return "", nil
	}
	if !telegramRe.MatchString(v) {
		return "", ErrInvalidTelegram
	}
	return v, nil
}

// ProfileView is a read model exposed to other users.
type ProfileView struct {
	UserID           UserID
	DisplayName      string
	Bio              string
	AvatarURL        *string
	TelegramUsername *string
	Visibility       Visibility
}
