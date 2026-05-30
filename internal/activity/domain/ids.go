package domain

import "github.com/google/uuid"

// ActivityID identifies a journal entry.
type ActivityID string

// UserID identifies a user.
type UserID string

// ParseActivityID parses UUID.
func ParseActivityID(raw string) (ActivityID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrNotFound
	}
	return ActivityID(raw), nil
}

// ParseUserID parses user UUID.
func ParseUserID(raw string) (UserID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrForbidden
	}
	return UserID(raw), nil
}
