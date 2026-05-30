package domain

// PasswordHash stores a bcrypt hash; never log or expose.
type PasswordHash string

// Credentials are authentication factors for a user.
type Credentials struct {
	UserID       UserID
	Email        Email
	PasswordHash PasswordHash
	Account      AccountState
}
