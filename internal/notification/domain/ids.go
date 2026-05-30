package domain

import (
	"strings"

	"github.com/google/uuid"
)

// NotificationID identifies an in-app notification.
type NotificationID string

// UserID identifies a platform user.
type UserID string

// ParseUserID validates UUID user id.
func ParseUserID(s string) (UserID, error) {
	s = strings.TrimSpace(s)
	if _, err := uuid.Parse(s); err != nil {
		return "", ErrValidation
	}
	return UserID(s), nil
}

// ParseNotificationID validates notification id.
func ParseNotificationID(s string) (NotificationID, error) {
	s = strings.TrimSpace(s)
	if _, err := uuid.Parse(s); err != nil {
		return "", ErrValidation
	}
	return NotificationID(s), nil
}
