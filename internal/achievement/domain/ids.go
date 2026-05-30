package domain

import "github.com/google/uuid"

// UserID identifies a user.
type UserID string

// AchievementCode identifies an achievement definition.
type AchievementCode string

// SourceEventID correlates grant with outbox event.
type SourceEventID string

// ParseUserID parses user UUID.
func ParseUserID(raw string) (UserID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrForbidden
	}
	return UserID(raw), nil
}

// ParseSourceEventID parses source event UUID.
func ParseSourceEventID(raw string) (SourceEventID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrInvalidRule
	}
	return SourceEventID(raw), nil
}
