package domain

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	Hash(password string) (PasswordHash, error)
	Verify(hash PasswordHash, password string) error
}
