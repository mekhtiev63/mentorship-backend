package domain

import "github.com/google/uuid"

// EventID identifies a calendar event.
type EventID string

// UserID identifies a user.
type UserID string

// ParseEventID parses event UUID.
func ParseEventID(raw string) (EventID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrNotFound
	}
	return EventID(raw), nil
}

// ParseUserID parses user UUID.
func ParseUserID(raw string) (UserID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrForbidden
	}
	return UserID(raw), nil
}
