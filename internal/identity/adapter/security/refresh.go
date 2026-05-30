package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// RefreshGenerator creates opaque refresh tokens.
type RefreshGenerator struct{}

// NewRefreshGenerator creates a RefreshGenerator.
func NewRefreshGenerator() *RefreshGenerator {
	return &RefreshGenerator{}
}

// Generate returns the raw token and SHA-256 hex hash for storage.
func (g *RefreshGenerator) Generate() (raw string, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("rand: %w", err)
	}
	raw = hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(sum[:])
	return raw, hash, nil
}
