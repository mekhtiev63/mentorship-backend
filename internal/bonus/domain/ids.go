package domain

import "github.com/google/uuid"

// ParseUserID parses user UUID.
func ParseUserID(raw string) (UserID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrNotFound
	}
	return UserID(raw), nil
}
