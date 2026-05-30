package domain

import "github.com/google/uuid"

// UserID is a canonical UUID string.
type UserID string

// ParseUserID validates and returns a UserID.
func ParseUserID(raw string) (UserID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrInvalidUserID
	}
	return UserID(raw), nil
}

// String returns the ID text.
func (id UserID) String() string {
	return string(id)
}

// IsZero reports whether the ID is empty.
func (id UserID) IsZero() bool {
	return id == ""
}
