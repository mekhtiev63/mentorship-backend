package security

import (
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/identity/domain"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// BcryptHasher hashes passwords with bcrypt.
type BcryptHasher struct{}

// NewBcryptHasher creates a BcryptHasher.
func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

// Hash returns a bcrypt hash.
func (h *BcryptHasher) Hash(password string) (domain.PasswordHash, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return domain.PasswordHash(bytes), nil
}

// Verify checks a password against a hash.
func (h *BcryptHasher) Verify(hash domain.PasswordHash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
