package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// UserID identifies a user account.
type UserID string

// ParseUserID validates UUID format.
func ParseUserID(raw string) (UserID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrNotFound
	}
	return UserID(raw), nil
}

// Email is a normalized email address.
type Email string

// ParseEmail validates email input.
func ParseEmail(raw string) (Email, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" || !strings.Contains(normalized, "@") {
		return "", ErrInvalidEmail
	}
	return Email(normalized), nil
}

// UserStatus is account status.
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

// ParseUserStatus parses status from string.
func ParseUserStatus(raw string) (UserStatus, error) {
	switch UserStatus(raw) {
	case UserStatusActive, UserStatusDisabled:
		return UserStatus(raw), nil
	default:
		return "", ErrInvalidStatus
	}
}

// User is the user account aggregate root (without password).
type User struct {
	ID        UserID
	Email     Email
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// IsDeleted reports soft deletion.
func (u User) IsDeleted() bool {
	return u.DeletedAt != nil
}
