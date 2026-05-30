package domain

import "strings"

// Email is a normalized email address.
type Email string

// ParseEmail normalizes an email for persistence and lookup.
func ParseEmail(raw string) (Email, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" || !strings.Contains(normalized, "@") {
		return "", ErrInvalidEmail
	}
	return Email(normalized), nil
}

// String returns the email text.
func (e Email) String() string {
	return string(e)
}
