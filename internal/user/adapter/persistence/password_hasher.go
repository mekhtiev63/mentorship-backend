package persistence

import (
	identitysecurity "github.com/go-mentorship-platform/backend/internal/identity/adapter/security"
)

// PasswordHasher adapts identity bcrypt hasher for user module.
type PasswordHasher struct {
	inner *identitysecurity.BcryptHasher
}

// NewPasswordHasher creates PasswordHasher.
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{inner: identitysecurity.NewBcryptHasher()}
}

// Hash implements domain.PasswordHasher.
func (h *PasswordHasher) Hash(password string) (string, error) {
	hash, err := h.inner.Hash(password)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
