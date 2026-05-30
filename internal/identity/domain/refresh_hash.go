package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HashRefreshToken returns the raw token and its stored hash.
func HashRefreshToken(raw string) (string, string, error) {
	if raw == "" {
		return "", "", fmt.Errorf("empty refresh token")
	}
	sum := sha256.Sum256([]byte(raw))
	return raw, hex.EncodeToString(sum[:]), nil
}
